package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_StatusHistory_Success_ReturnsStatusHistory(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfStatusHistory:123",
			"code": "SH-001",
			"parent_id": "CmfTask:456",
			"old_status": "OPEN",
			"new_status": "IN_PROGRESS"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfStatusHistory.get")
	}

	history, meta, err := client.StatusHistory(testCtx, "CmfStatusHistory:123", nil)

	require.NoError(t, err)
	require.NotNil(t, history)
	assert.Equal(t, "CmfStatusHistory:123", history.ID)
	assert.Equal(t, "SH-001", history.Code)
	require.NotNil(t, history.OldStatus)
	assert.Equal(t, "OPEN", *history.OldStatus)
	require.NotNil(t, history.NewStatus)
	assert.Equal(t, "IN_PROGRESS", *history.NewStatus)
	assert.NotNil(t, meta)
}

func TestClient_StatusHistory_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "StatusHistory not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	history, meta, err := client.StatusHistory(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "StatusHistory not found")
	assert.Nil(t, history)
	assert.Nil(t, meta)
}

func TestClient_StatusHistory_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	history, meta, err := client.StatusHistory(testCtx, "CmfStatusHistory:123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, history)
	assert.Nil(t, meta)
}

func TestClient_StatusHistoryQuery_Success_ReturnsStatusHistory(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfStatusHistory:123",
			"code": "SH-001",
			"old_status": "OPEN",
			"new_status": "IN_PROGRESS"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfStatusHistory.get")
	}

	qb := NewQueryBuilder().From(EntityStatusHistory)
	history, meta, err := client.StatusHistoryQuery(testCtx, qb)

	require.NoError(t, err)
	require.NotNil(t, history)
	assert.Equal(t, "CmfStatusHistory:123", history.ID)
	assert.NotNil(t, meta)
}

func TestClient_StatusHistoryList_Success_ReturnsStatusHistories(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfStatusHistory:1", "code": "SH-001", "old_status": "OPEN", "new_status": "IN_PROGRESS"},
			{"id": "CmfStatusHistory:2", "code": "SH-002", "old_status": "IN_PROGRESS", "new_status": "DONE"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfStatusHistory.list")
	}

	qb := NewQueryBuilder().From(EntityStatusHistory)
	histories, meta, err := client.StatusHistoryList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, histories, 2)
	require.NotNil(t, histories[0].OldStatus)
	assert.Equal(t, "OPEN", *histories[0].OldStatus)
	require.NotNil(t, histories[1].NewStatus)
	assert.Equal(t, "DONE", *histories[1].NewStatus)
	assert.NotNil(t, meta)
}

func TestClient_StatusHistoryCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 42
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfStatusHistory.count")
	}

	qb := NewQueryBuilder().From(EntityStatusHistory)
	count, err := client.StatusHistoryCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 42, count)
}

func TestClient_TaskStatusHistory_Success_ReturnsHistories(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfStatusHistory:1", "parent_id": "CmfTask:123", "old_status": "OPEN", "new_status": "IN_PROGRESS"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	histories, meta, err := client.TaskStatusHistory(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.NotNil(t, meta)
}

func TestClient_ProjectStatusHistory_Success_ReturnsHistories(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfStatusHistory:1", "project_id": "CmfProject:123", "old_status": "OPEN", "new_status": "CLOSED"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	histories, meta, err := client.ProjectStatusHistory(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.NotNil(t, meta)
}

func TestClient_StatusHistories_Success_ReturnsHistories(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfStatusHistory:1", "code": "SH-001"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfStatusHistory.list")
	}

	histories, meta, err := client.StatusHistories(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.NotNil(t, meta)
}
