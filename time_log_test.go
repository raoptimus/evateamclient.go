/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	encjson "encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TimeLog_Success_ReturnsTimeLog(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTimeTrackerHistory:123",
			"time_spent": 180,
			"parent_id": "CmfTask:456"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.get")
	}

	log, meta, err := client.TimeLog(testCtx, "CmfTimeTrackerHistory:123", nil)

	require.NoError(t, err)
	require.NotNil(t, log)
	assert.Equal(t, "CmfTimeTrackerHistory:123", log.ID)
	assert.Equal(t, 180, log.TimeSpent)
	assert.NotNil(t, meta)
}

func TestClient_TimeLog_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "TimeLog not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	log, meta, err := client.TimeLog(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TimeLog not found")
	assert.Nil(t, log)
	assert.Nil(t, meta)
}

func TestClient_TimeLog_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	log, meta, err := client.TimeLog(testCtx, "CmfTimeTrackerHistory:123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, log)
	assert.Nil(t, meta)
}

func TestClient_TimeLogsList_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 60},
			{"id": "TTH:2", "time_spent": 120}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.list")
	}

	qb := NewQueryBuilder().From(EntityTimeLog)
	logs, meta, err := client.TimeLogsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, 60, logs[0].TimeSpent)
	assert.Equal(t, 120, logs[1].TimeSpent)
	assert.NotNil(t, meta)
}

func TestClient_TimeLogCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 75
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.count")
	}

	qb := NewQueryBuilder().From(EntityTimeLog)
	count, err := client.TimeLogCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 75, count)
}

func TestClient_TaskTimeLogs_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 60, "parent_id": "CmfTask:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	logs, meta, err := client.TaskTimeLogs(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, meta)
}

func TestClient_UserTimeLogs_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 120, "cmf_owner_id": "CmfPerson:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	logs, meta, err := client.UserTimeLogs(testCtx, "CmfPerson:123", nil)

	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, meta)
}

func TestClient_UserTaskTimeLogs_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 90}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	logs, meta, err := client.UserTaskTimeLogs(testCtx, "CmfTask:123", "CmfPerson:456", nil)

	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, meta)
}

func TestClient_ProjectTimeLogs_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 60}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	logs, meta, err := client.ProjectTimeLogs(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, meta)
}

func TestClient_TimeLogs_Success_ReturnsLogs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "TTH:1", "time_spent": 60}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.list")
	}

	logs, meta, err := client.TimeLogs(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, logs, 1)
	assert.NotNil(t, meta)
}

func TestClient_TimeLogCreate_Success_ReturnsTimeLog(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTimeTrackerHistory:new-123",
			"time_spent": 180,
			"parent_id": "CmfTask:456"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.create")
	}

	params := TimeLogCreateParams{
		ParentID:  "CmfTask:456",
		TimeSpent: 180,
	}
	log, err := client.TimeLogCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfTimeTrackerHistory:new-123", log.ID)
	assert.Equal(t, 180, log.TimeSpent)
}

func TestClient_TimeLogCreate_ResultIsIDString_FetchesCreatedTimeLog(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTimeTrackerHistory:new-123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfTimeTrackerHistory:new-123", "time_spent": 180, "parent_id": "CmfTask:456"}
		}`),
	}

	params := TimeLogCreateParams{ParentID: "CmfTask:456", TimeSpent: 180}
	log, err := client.TimeLogCreate(testCtx, params)

	require.NoError(t, err)
	require.NotNil(t, log)
	assert.Equal(t, "CmfTimeTrackerHistory:new-123", log.ID)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + follow-up get")
}

// TestClient_TimeLogCreate_EmptyResult_ReturnsError guards against a silent
// failure (empty/null/false result) being mistaken for a zero-value time log.
func TestClient_TimeLogCreate_EmptyResult_ReturnsError(t *testing.T) {
	tests := []struct {
		name   string
		result string
	}{
		{"null", `null`},
		{"empty string", `""`},
		{"false", `false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClient(t)
			mockHTTP.response = mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":`+tt.result+`}`)

			params := TimeLogCreateParams{ParentID: "CmfTask:456", TimeSpent: 180}
			log, err := client.TimeLogCreate(testCtx, params)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CmfTimeTrackerHistory.create returned empty result")
			assert.Nil(t, log)
		})
	}
}

// TestClient_TimeLogUpdate_Success_SendsIDInArgs is the regression test for
// the EVA update convention: the id belongs in Args, not Kwargs (see TaskUpdate).
func TestClient_TimeLogUpdate_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTimeTrackerHistory:123",
			"time_spent": 240
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.update")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfTimeTrackerHistory:123"}, parsed.Args) &&
			assert.Equal(t, map[string]any{"time_spent": float64(240)}, parsed.Kwargs)
	}

	updates := map[string]any{"time_spent": 240}
	log, err := client.TimeLogUpdate(testCtx, "CmfTimeTrackerHistory:123", updates)

	require.NoError(t, err)
	assert.Equal(t, 240, log.TimeSpent)
}

func TestClient_TimeLogUpdate_ResultIsIDString_FetchesUpdatedTimeLog(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfTimeTrackerHistory:123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfTimeTrackerHistory:123", "time_spent": 240}
		}`),
	}

	updates := map[string]any{"time_spent": 240}
	log, err := client.TimeLogUpdate(testCtx, "CmfTimeTrackerHistory:123", updates)

	require.NoError(t, err)
	require.NotNil(t, log)
	assert.Equal(t, 240, log.TimeSpent)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected update + follow-up get")
}

func TestClient_TimeLogUpdate_EmptyTimeLogID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	log, err := client.TimeLogUpdate(testCtx, "", map[string]any{"time_spent": 240})

	require.Error(t, err)
	assert.Nil(t, log)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}

// TestClient_TimeLogDelete_Success_SendsIDInArgs is the regression test for
// the EVA delete convention: the id belongs in Args, not Kwargs (see TaskDelete).
func TestClient_TimeLogDelete_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTimeTrackerHistory.delete")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfTimeTrackerHistory:123"}, parsed.Args) &&
			assert.Nil(t, parsed.Kwargs)
	}

	err := client.TimeLogDelete(testCtx, "CmfTimeTrackerHistory:123")

	assert.NoError(t, err)
}

func TestClient_TimeLogDelete_EmptyTimeLogID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	err := client.TimeLogDelete(testCtx, "")

	assert.Error(t, err)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}
