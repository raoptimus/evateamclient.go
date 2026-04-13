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
	"net/http"
	"strings"
	"testing"

	"github.com/imroc/req/v3"
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

func TestClient_TimeSpentStats_Success_ReturnsGroupedReport(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	// Response 1: TimeLogs list (2 logs, 2 persons, 2 tasks)
	timeLogsResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfTimeTrackerHistory:log1",
				"time_spent": 120,
				"cmf_owner_id": "CmfPerson:person1",
				"parent_id": "CmfTask:task1",
				"parent": {"id": "CmfTask:task1", "code": "PRJ-001", "name": "Task One"}
			},
			{
				"id": "CmfTimeTrackerHistory:log2",
				"time_spent": 60,
				"cmf_owner_id": "CmfPerson:person2",
				"parent_id": "CmfTask:task2",
				"parent": {"id": "CmfTask:task2", "code": "PRJ-002", "name": "Task Two"}
			},
			{
				"id": "CmfTimeTrackerHistory:log3",
				"time_spent": 30,
				"cmf_owner_id": "CmfPerson:person1",
				"parent_id": "CmfTask:task2",
				"parent": {"id": "CmfTask:task2", "code": "PRJ-002", "name": "Task Two"}
			}
		],
		"meta": {"total": 3}
	}`

	// Response 2: Person 1
	person1Resp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:person1", "name": "Alice"},
		"meta": {}
	}`

	// Response 3: Person 2
	person2Resp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:person2", "name": "Bob"},
		"meta": {}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, timeLogsResp),
		mockResponse(http.StatusOK, person1Resp),
		mockResponse(http.StatusOK, person2Resp),
	}

	stats, err := client.TimeSpentStats(testCtx, TimeSpentStatsParams{
		ProjectID: "CmfProject:proj1",
	})

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, "CmfProject:proj1", stats.ProjectID)
	assert.Equal(t, 210, stats.GrandTotalTime)
	assert.Len(t, stats.Persons, 2)

	// Persons sorted by TotalTime DESC: Alice (150) > Bob (60)
	assert.Equal(t, "Alice", stats.Persons[0].PersonName)
	assert.Equal(t, 150, stats.Persons[0].TotalTime)
	assert.Len(t, stats.Persons[0].Tasks, 2)

	// Alice's tasks sorted by TimeSpent DESC: Task One (120) > Task Two (30)
	assert.Equal(t, "PRJ-001", stats.Persons[0].Tasks[0].TaskCode)
	assert.Equal(t, 120, stats.Persons[0].Tasks[0].TimeSpent)
	assert.Equal(t, "PRJ-002", stats.Persons[0].Tasks[1].TaskCode)
	assert.Equal(t, 30, stats.Persons[0].Tasks[1].TimeSpent)

	assert.Equal(t, "Bob", stats.Persons[1].PersonName)
	assert.Equal(t, 60, stats.Persons[1].TotalTime)
	assert.Len(t, stats.Persons[1].Tasks, 1)
}

func TestClient_TimeSpentStats_EmptyResult_ReturnsEmptyReport(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	emptyResp := `{
		"jsonrpc": "2.2",
		"result": [],
		"meta": {"total": 0}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, emptyResp),
	}

	stats, err := client.TimeSpentStats(testCtx, TimeSpentStatsParams{
		ProjectID: "CmfProject:proj1",
	})

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, "CmfProject:proj1", stats.ProjectID)
	assert.Empty(t, stats.Persons)
	assert.Equal(t, 0, stats.GrandTotalTime)
}

func TestClient_TimeSpentStats_TimeLogsFetchError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.errors = []error{errors.New("connection refused")}
	mockHTTP.responses = []*req.Response{nil}

	stats, err := client.TimeSpentStats(testCtx, TimeSpentStatsParams{
		ProjectID: "CmfProject:proj1",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, stats)
}

func TestClient_TimeSpentStats_WithDateFilter_PassesDatesToAPI(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	timeLogsResp := `{
		"jsonrpc": "2.2",
		"result": [],
		"meta": {"total": 0}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, timeLogsResp),
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		s := string(body)
		return assert.Contains(t, s, "2025-01-01") &&
			assert.Contains(t, s, "2025-12-31")
	}

	stats, err := client.TimeSpentStats(testCtx, TimeSpentStatsParams{
		ProjectID: "CmfProject:proj1",
		DateFrom:  "2025-01-01",
		DateTo:    "2025-12-31",
	})

	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, "2025-01-01", stats.DateFrom)
	assert.Equal(t, "2025-12-31", stats.DateTo)
}

func TestClient_SprintExecutorsKPI_EmptySprintCode_UsesSprintPrefixWildcard(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	projectResp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfProject:proj1"},
		"meta": {}
	}`

	emptyListsResp := `{
		"jsonrpc": "2.2",
		"result": [],
		"meta": {"total": 0}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, projectResp),
		mockResponse(http.StatusOK, emptyListsResp),
	}

	callNum := 0
	mockHTTP.bodyCheck = func(body []byte) bool {
		callNum++
		if callNum != 2 {
			return true
		}
		return strings.Contains(string(body), "SPR-%")
	}

	report, err := client.SprintExecutorsKPI(testCtx, &SprintExecutorsKPIParams{
		ProjectCode: "epud",
	})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, 0, report.BaselineTasks)
	assert.Equal(t, 0, report.TotalClosedTasks)
	assert.Empty(t, report.Executors)
}
