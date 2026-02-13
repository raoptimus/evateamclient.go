package evateamclient

import (
	"errors"
	"net/http"
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

func TestClient_SprintExecutorsKPI_FirstInProgress_ExcludesTasksCreatedDuringSprint(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	sprintResp := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:spr1",
			"code": "SPR-001",
			"project_id": "CmfProject:epud",
			"start_date": "2025-01-10",
			"end_date": "2025-01-20"
		},
		"meta": {}
	}`

	tasksResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfTask:1",
				"code": "EPUD-1",
				"name": "Task 1",
				"lists": [{"id": "CmfList:spr1", "code": "SPR-001"}],
				"cmf_created_at": "2025-01-01T09:00:00+03:00",
				"status_closed_at": "2025-01-15T10:00:00+03:00"
			},
			{
				"id": "CmfTask:2",
				"code": "EPUD-2",
				"name": "Task 2",
				"lists": [{"id": "CmfList:spr1", "code": "SPR-001"}],
				"cmf_created_at": "2025-01-12T09:00:00+03:00",
				"status_closed_at": "2025-01-18T12:00:00+03:00"
			},
			{
				"id": "CmfTask:3",
				"code": "EPUD-3",
				"name": "Task 3",
				"lists": [{"id": "CmfList:spr1", "code": "SPR-001"}],
				"cmf_created_at": "2025-01-05T09:00:00+03:00"
			}
		],
		"meta": {"total": 3}
	}`

	statusHistoryResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfStatusHistory:1",
				"parent_id": "CmfTask:1",
				"new_status": "IN_PROGRESS",
				"cmf_owner_id": "CmfPerson:p1",
				"cmf_created_at": "2025-01-11T09:30:00+03:00"
			}
		],
		"meta": {"total": 1}
	}`

	personResp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:p1", "name": "Alice"},
		"meta": {}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, sprintResp),
		mockResponse(http.StatusOK, tasksResp),
		mockResponse(http.StatusOK, statusHistoryResp),
		mockResponse(http.StatusOK, personResp),
	}

	report, err := client.SprintExecutorsKPI(testCtx, SprintExecutorsKPIParams{
		ProjectID:    "CmfProject:epud",
		SprintCode:   "SPR-001",
		AssigneeMode: SprintKPIAssigneeModeFirstInProgress,
	})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, "CmfProject:epud", report.ProjectID)
	assert.Equal(t, "SPR-001", report.SprintCode)
	assert.Equal(t, 2, report.BaselineTasks)
	assert.Equal(t, 1, report.ExcludedNewTasks)
	assert.Equal(t, 1, report.TotalClosedTasks)
	assert.Equal(t, 0, report.UnassignedClosed)
	require.Len(t, report.Executors, 1)
	assert.Equal(t, "CmfPerson:p1", report.Executors[0].PersonID)
	assert.Equal(t, "Alice", report.Executors[0].PersonName)
	assert.Equal(t, 1, report.Executors[0].ClosedTasks)
	assert.Equal(t, []string{"EPUD-1"}, report.Executors[0].TaskCodes)
}

func TestClient_SprintExecutorsKPI_MaxTimeSpent_AssignsByTopLogger(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	sprintResp := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:spr2",
			"code": "SPR-002",
			"project_id": "CmfProject:epud",
			"start_date": "2025-02-01",
			"end_date": "2025-02-14"
		},
		"meta": {}
	}`

	tasksResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfTask:11",
				"code": "EPUD-11",
				"name": "Task 11",
				"lists": [{"id": "CmfList:spr2", "code": "SPR-002"}],
				"cmf_created_at": "2025-01-20T09:00:00+03:00",
				"status_closed_at": "2025-02-10T10:00:00+03:00"
			},
			{
				"id": "CmfTask:12",
				"code": "EPUD-12",
				"name": "Task 12",
				"lists": [{"id": "CmfList:spr2", "code": "SPR-002"}],
				"cmf_created_at": "2025-01-25T09:00:00+03:00",
				"status_closed_at": "2025-02-13T10:00:00+03:00"
			}
		],
		"meta": {"total": 2}
	}`

	logsTask11Resp := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "log-11-1", "time_spent": 60, "cmf_owner_id": "CmfPerson:p1"},
			{"id": "log-11-2", "time_spent": 120, "cmf_owner_id": "CmfPerson:p2"}
		],
		"meta": {"total": 2}
	}`

	logsTask12Resp := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "log-12-1", "time_spent": 30, "cmf_owner_id": "CmfPerson:p1"},
			{"id": "log-12-2", "time_spent": 30, "cmf_owner_id": "CmfPerson:p2"}
		],
		"meta": {"total": 2}
	}`

	person1Resp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:p1", "name": "Alice"},
		"meta": {}
	}`

	person2Resp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:p2", "name": "Bob"},
		"meta": {}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, sprintResp),
		mockResponse(http.StatusOK, tasksResp),
		mockResponse(http.StatusOK, logsTask11Resp),
		mockResponse(http.StatusOK, logsTask12Resp),
		mockResponse(http.StatusOK, person1Resp),
		mockResponse(http.StatusOK, person2Resp),
	}

	report, err := client.SprintExecutorsKPI(testCtx, SprintExecutorsKPIParams{
		ProjectID:    "CmfProject:epud",
		SprintCode:   "SPR-002",
		AssigneeMode: SprintKPIAssigneeModeMaxTimeSpent,
	})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, 2, report.BaselineTasks)
	assert.Equal(t, 0, report.ExcludedNewTasks)
	assert.Equal(t, 2, report.TotalClosedTasks)
	assert.Equal(t, 0, report.UnassignedClosed)
	require.Len(t, report.Executors, 2)

	assert.Equal(t, "CmfPerson:p1", report.Executors[0].PersonID)
	assert.Equal(t, "Alice", report.Executors[0].PersonName)
	assert.Equal(t, 1, report.Executors[0].ClosedTasks)
	assert.Equal(t, []string{"EPUD-12"}, report.Executors[0].TaskCodes)

	assert.Equal(t, "CmfPerson:p2", report.Executors[1].PersonID)
	assert.Equal(t, "Bob", report.Executors[1].PersonName)
	assert.Equal(t, 1, report.Executors[1].ClosedTasks)
	assert.Equal(t, []string{"EPUD-11"}, report.Executors[1].TaskCodes)
}

func TestClient_SprintExecutorsKPI_MissingSprintDates_DoesNotFail(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	sprintResp := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:spr3",
			"code": "SPR-003",
			"project_id": "CmfProject:epud"
		},
		"meta": {}
	}`

	tasksResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfTask:21",
				"code": "EPUD-21",
				"name": "Task 21",
				"lists": [{"id": "CmfList:spr3", "code": "SPR-003"}],
				"cmf_created_at": "2025-03-01T09:00:00+03:00",
				"status_closed_at": "2025-03-03T10:00:00+03:00"
			},
			{
				"id": "CmfTask:22",
				"code": "EPUD-22",
				"name": "Task 22",
				"lists": [{"id": "CmfList:spr3", "code": "SPR-003"}],
				"cmf_created_at": "2025-03-02T09:00:00+03:00"
			}
		],
		"meta": {"total": 2}
	}`

	statusHistoryResp := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfStatusHistory:21",
				"parent_id": "CmfTask:21",
				"new_status": "IN_PROGRESS",
				"cmf_owner_id": "CmfPerson:p1",
				"cmf_created_at": "2025-03-02T09:30:00+03:00"
			}
		],
		"meta": {"total": 1}
	}`

	personResp := `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfPerson:p1", "name": "Alice"},
		"meta": {}
	}`

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, sprintResp),
		mockResponse(http.StatusOK, tasksResp),
		mockResponse(http.StatusOK, statusHistoryResp),
		mockResponse(http.StatusOK, personResp),
	}

	report, err := client.SprintExecutorsKPI(testCtx, SprintExecutorsKPIParams{
		ProjectID:    "CmfProject:epud",
		SprintCode:   "SPR-003",
		AssigneeMode: SprintKPIAssigneeModeFirstInProgress,
	})

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.Equal(t, "CmfProject:epud", report.ProjectID)
	assert.Equal(t, "SPR-003", report.SprintCode)
	assert.Equal(t, "", report.SprintStartDate)
	assert.Equal(t, "", report.SprintEndDate)
	assert.Equal(t, 2, report.BaselineTasks)
	assert.Equal(t, 0, report.ExcludedNewTasks)
	assert.Equal(t, 1, report.TotalClosedTasks)
	assert.Equal(t, 0, report.UnassignedClosed)
	require.Len(t, report.Executors, 1)
	assert.Equal(t, "CmfPerson:p1", report.Executors[0].PersonID)
	assert.Equal(t, "Alice", report.Executors[0].PersonName)
	assert.Equal(t, 1, report.Executors[0].ClosedTasks)
	assert.Equal(t, []string{"EPUD-21"}, report.Executors[0].TaskCodes)
}
