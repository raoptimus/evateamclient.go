/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturedRequest is a minimal view of a JSON-RPC request body used by the
// status-update tests to inspect method and kwargs of each call.
type capturedRequest struct {
	Method string         `json:"method"`
	Kwargs map[string]any `json:"kwargs"`
}

// captureBodies installs a bodyCheck that records every request body parsed as
// a capturedRequest. It always returns true so requests are never rejected.
func captureBodies(mock *sequentialMockHTTPClient, sink *[]capturedRequest) {
	mock.bodyCheck = func(body []byte) bool {
		var req capturedRequest
		_ = json.Unmarshal(body, &req)
		*sink = append(*sink, req)
		return true
	}
}

func taskGetResp(id, epicID string) string {
	return fmt.Sprintf(`{"jsonrpc":"2.2","result":{"id":%q,"epic_id":%q},"meta":{"total":1}}`, id, epicID)
}

func countMethod(reqs []capturedRequest, method string) int {
	n := 0
	for _, r := range reqs {
		if r.Method == method {
			n++
		}
	}
	return n
}

// When the status transition resets the epic, it must be restored with an
// explicit {"epic": <original>} update and the returned task keeps the epic.
func TestClient_TaskUpdateStatus_EpicReset_RestoresEpic(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)
	var reqs []capturedRequest
	captureBodies(mockHTTP, &reqs)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "CmfTask:E1")),   // pre-read: epic present
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTask:T1"}`), // status update
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "")),             // re-fetch: epic reset
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTask:T1"}`), // restore update
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "CmfTask:E1")),   // re-fetch: epic restored
	}

	task, err := client.TaskUpdateStatus(testCtx, "CmfTask:T1", "Backlog")

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "CmfTask:E1", task.EpicID, "epic must be restored")
	assert.Equal(t, 5, mockHTTP.callIdx, "expected get+update+get+update+get")

	require.Equal(t, 2, countMethod(reqs, "CmfTask.update"), "expected status update + restore update")
	// The second update is the restore and must carry the original epic.
	var updates []capturedRequest
	for _, r := range reqs {
		if r.Method == "CmfTask.update" {
			updates = append(updates, r)
		}
	}
	require.Len(t, updates, 2)
	assert.Equal(t, "CmfTask:E1", updates[1].Kwargs["epic"], "restore update must set epic to original")
}

// A task without an epic must not get one assigned, even though the post-update
// epic is empty too. Exactly one update call (the status change) is expected.
func TestClient_TaskUpdateStatus_NoEpic_NoRestore(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)
	var reqs []capturedRequest
	captureBodies(mockHTTP, &reqs)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "")),             // pre-read: no epic
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTask:T1"}`), // status update
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "")),             // re-fetch: still no epic
	}

	task, err := client.TaskUpdateStatus(testCtx, "CmfTask:T1", "Backlog")

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Empty(t, task.EpicID)
	assert.Equal(t, 3, mockHTTP.callIdx, "expected get+update+get, no restore")
	assert.Equal(t, 1, countMethod(reqs, "CmfTask.update"), "no restore update expected")
}

// When the transition leaves the epic unchanged, no restore update is issued.
func TestClient_TaskUpdateStatus_EpicUnchanged_NoRestore(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)
	var reqs []capturedRequest
	captureBodies(mockHTTP, &reqs)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "CmfTask:E1")),   // pre-read
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTask:T1"}`), // status update
		mockResponse(http.StatusOK, taskGetResp("CmfTask:T1", "CmfTask:E1")),   // re-fetch: unchanged
	}

	task, err := client.TaskUpdateStatus(testCtx, "CmfTask:T1", "IN_PROGRESS")

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "CmfTask:E1", task.EpicID)
	assert.Equal(t, 3, mockHTTP.callIdx, "expected get+update+get, no restore")
	assert.Equal(t, 1, countMethod(reqs, "CmfTask.update"))
}

// If the pre-read fails, the status update must not be attempted and the error
// is surfaced (no silent loss of epic preservation).
func TestClient_TaskUpdateStatus_PreReadFails_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)
	var reqs []capturedRequest
	captureBodies(mockHTTP, &reqs)

	mockHTTP.errors = []error{errors.New("connection refused")}

	task, err := client.TaskUpdateStatus(testCtx, "CmfTask:T1", "Backlog")

	require.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, 1, mockHTTP.callIdx, "status update must not run after a failed pre-read")
	assert.Equal(t, 0, countMethod(reqs, "CmfTask.update"))
}

func TestClient_TaskUpdateStatus_EmptyID_ReturnsError(t *testing.T) {
	client, _ := newTestClientWithSequentialMock(t)

	task, err := client.TaskUpdateStatus(testCtx, "", "Backlog")

	require.Error(t, err)
	assert.Nil(t, task)
}
