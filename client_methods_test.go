package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Task_Success_ReturnsTask(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTask:123",
			"code": "TASK-001",
			"name": "Test Task",
			"cache_status_type": "OPEN"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.get")
	}

	task, meta, err := client.Task(testCtx, "TASK-001", nil)

	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "CmfTask:123", task.ID)
	assert.Equal(t, "TASK-001", task.Code)
	assert.Equal(t, "Test Task", task.Name)
	assert.NotNil(t, meta)
}

func TestClient_Task_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Task not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	task, meta, err := client.Task(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Task not found")
	assert.Nil(t, task)
	assert.Nil(t, meta)
}

func TestClient_Task_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	task, meta, err := client.Task(testCtx, "TASK-001", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, task)
	assert.Nil(t, meta)
}

func TestClient_TasksList_Success_ReturnsTasks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTask:1", "code": "TASK-001", "name": "Task 1"},
			{"id": "CmfTask:2", "code": "TASK-002", "name": "Task 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.list")
	}

	qb := NewQueryBuilder().From(EntityTask)
	tasks, meta, err := client.TasksList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "TASK-001", tasks[0].Code)
	assert.Equal(t, "TASK-002", tasks[1].Code)
	assert.NotNil(t, meta)
}

func TestClient_TaskCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 42
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.count")
	}

	qb := NewQueryBuilder().From(EntityTask)
	count, err := client.TaskCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 42, count)
}

func TestClient_TaskCreate_Success_ReturnsTask(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTask:new-123",
			"code": "TASK-100",
			"name": "New Task"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New Task")
	}

	params := TaskCreateParams{
		Name:      "New Task",
		ProjectID: "Project:123",
	}
	task, err := client.TaskCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfTask:new-123", task.ID)
	assert.Equal(t, "New Task", task.Name)
}

func TestClient_TaskUpdate_Success_ReturnsUpdatedTask(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfTask:123",
			"code": "TASK-001",
			"name": "Updated Name"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.update")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "Updated Name")
	}

	updates := map[string]any{"name": "Updated Name"}
	task, err := client.TaskUpdate(testCtx, "CmfTask:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", task.Name)
}

func TestClient_TaskDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTask.delete")
	}

	err := client.TaskDelete(testCtx, "CmfTask:123")

	assert.NoError(t, err)
}
