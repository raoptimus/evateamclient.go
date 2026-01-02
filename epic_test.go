package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Epic_Success_ReturnsEpic(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTask:123",
			"code": "UDMP-123",
			"name": "Test Epic",
			"logic_type": {"code": "task.epic"}
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.get")
	}

	epic, meta, err := client.Epic(testCtx, "UDMP-123", nil)

	require.NoError(t, err)
	require.NotNil(t, epic)
	assert.Equal(t, "CmfTask:123", epic.ID)
	assert.Equal(t, "UDMP-123", epic.Code)
	assert.Equal(t, "Test Epic", epic.Name)
	assert.NotNil(t, meta)
}

func TestClient_Epic_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Epic not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	epic, meta, err := client.Epic(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Epic not found")
	assert.Nil(t, epic)
	assert.Nil(t, meta)
}

func TestClient_Epic_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	epic, meta, err := client.Epic(testCtx, "UDMP-123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, epic)
	assert.Nil(t, meta)
}

func TestClient_EpicByID_Success_ReturnsEpic(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTask:123",
			"code": "UDMP-123",
			"name": "Test Epic"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.get")
	}

	epic, meta, err := client.EpicByID(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	require.NotNil(t, epic)
	assert.Equal(t, "CmfTask:123", epic.ID)
	assert.NotNil(t, meta)
}

func TestClient_ProjectEpics_Success_ReturnsEpics(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTask:1", "code": "UDMP-001", "name": "Epic 1"},
			{"id": "CmfTask:2", "code": "UDMP-002", "name": "Epic 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.list")
	}

	epics, meta, err := client.ProjectEpics(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, epics, 2)
	assert.NotNil(t, meta)
}

func TestClient_EpicTasks_Success_ReturnsTasks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTask:1", "code": "PROJ-001", "epic_id": "CmfTask:123"},
			{"id": "CmfTask:2", "code": "PROJ-002", "epic_id": "CmfTask:123"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.list")
	}

	tasks, meta, err := client.EpicTasks(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.NotNil(t, meta)
}

func TestClient_Epics_Success_ReturnsEpics(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTask:1", "code": "UDMP-001", "name": "Epic 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.list")
	}

	epics, meta, err := client.Epics(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.NotNil(t, meta)
}
