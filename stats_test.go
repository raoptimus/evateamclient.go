package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TasksCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 150,
		"meta": {"total": 150}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}

	count, meta, err := client.TasksCount(testCtx, nil)

	require.NoError(t, err)
	assert.Equal(t, int64(150), count)
	assert.NotNil(t, meta)
}

func TestClient_TasksCount_WithFilters_ReturnsFilteredCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 42,
		"meta": {"total": 42}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}

	kwargs := map[string]any{
		"filter": []any{"project_id", "==", "CmfProject:123"},
	}
	count, meta, err := client.TasksCount(testCtx, kwargs)

	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
	assert.NotNil(t, meta)
}

func TestClient_TasksCount_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	count, meta, err := client.TasksCount(testCtx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Equal(t, int64(0), count)
	assert.Nil(t, meta)
}

func TestClient_ProjectTasksCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 75,
		"meta": {"total": 75}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "CmfProject:123")
	}

	count, meta, err := client.ProjectTasksCount(testCtx, "CmfProject:123")

	require.NoError(t, err)
	assert.Equal(t, int64(75), count)
	assert.NotNil(t, meta)
}

func TestClient_SprintTasksCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 25,
		"meta": {"total": 25}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "SPR-001")
	}

	count, meta, err := client.SprintTasksCount(testCtx, "SPR-001")

	require.NoError(t, err)
	assert.Equal(t, int64(25), count)
	assert.NotNil(t, meta)
}

func TestClient_ListTasksCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 30,
		"meta": {"total": 30}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "LIST-001")
	}

	count, meta, err := client.ListTasksCount(testCtx, "LIST-001")

	require.NoError(t, err)
	assert.Equal(t, int64(30), count)
	assert.NotNil(t, meta)
}
