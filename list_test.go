package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_List_Success_ReturnsList(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:123",
			"code": "SPR-001543",
			"name": "Sprint 1",
			"cache_status_type": "OPEN"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.get")
	}

	list, meta, err := client.List(testCtx, "SPR-001543", nil)

	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Equal(t, "CmfList:123", list.ID)
	assert.Equal(t, "SPR-001543", list.Code)
	assert.Equal(t, "Sprint 1", list.Name)
	assert.NotNil(t, meta)
}

func TestClient_List_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "List not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	list, meta, err := client.List(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "List not found")
	assert.Nil(t, list)
	assert.Nil(t, meta)
}

func TestClient_List_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	list, meta, err := client.List(testCtx, "SPR-001543", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, list)
	assert.Nil(t, meta)
}

func TestClient_ListsList_Success_ReturnsLists(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "name": "Sprint 1"},
			{"id": "CmfList:2", "code": "SPR-002", "name": "Sprint 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.list")
	}

	qb := NewQueryBuilder().From(EntityList)
	lists, meta, err := client.ListsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, lists, 2)
	assert.Equal(t, "SPR-001", lists[0].Code)
	assert.Equal(t, "SPR-002", lists[1].Code)
	assert.NotNil(t, meta)
}

func TestClient_ListCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 15
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.count")
	}

	qb := NewQueryBuilder().From(EntityList)
	count, err := client.ListCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 15, count)
}

func TestClient_Lists_Success_ReturnsLists(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "name": "Sprint 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.list")
	}

	lists, meta, err := client.Lists(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, lists, 1)
	assert.NotNil(t, meta)
}

func TestClient_ProjectLists_Success_ReturnsLists(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "project_id": "CmfProject:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	lists, meta, err := client.ProjectLists(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, lists, 1)
	assert.NotNil(t, meta)
}

func TestClient_OpenProjectLists_Success_ReturnsOpenLists(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "cache_status_type": "OPEN"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	lists, meta, err := client.OpenProjectLists(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, lists, 1)
	assert.NotNil(t, meta)
}

func TestClient_ListCreate_Success_ReturnsList(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:new-123",
			"code": "SPR-100",
			"name": "New Sprint"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New Sprint")
	}

	params := ListCreateParams{
		Name:     "New Sprint",
		ParentID: "CmfProject:123",
	}
	list, err := client.ListCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfList:new-123", list.ID)
	assert.Equal(t, "New Sprint", list.Name)
}

func TestClient_ListUpdate_Success_ReturnsUpdatedList(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:123",
			"code": "SPR-001",
			"name": "Updated Sprint"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.update")
	}

	updates := map[string]any{"name": "Updated Sprint"}
	list, err := client.ListUpdate(testCtx, "CmfList:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Sprint", list.Name)
}

func TestClient_ListClose_Success_ReturnsClosedList(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfList:123",
			"cache_status_type": "CLOSED"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.update")
	}

	list, err := client.ListClose(testCtx, "CmfList:123")

	require.NoError(t, err)
	assert.Equal(t, "CLOSED", list.CacheStatusType)
}

func TestClient_ListDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfList.delete")
	}

	err := client.ListDelete(testCtx, "CmfList:123")

	assert.NoError(t, err)
}

func TestClient_ProjectSprints_Success_ReturnsSprints(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "name": "Sprint 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	sprints, meta, err := client.ProjectSprints(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, sprints, 1)
	assert.NotNil(t, meta)
}

func TestClient_OpenProjectSprints_Success_ReturnsOpenSprints(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "SPR-001", "cache_status_type": "OPEN"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	sprints, meta, err := client.OpenProjectSprints(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, sprints, 1)
	assert.NotNil(t, meta)
}

func TestClient_ProjectReleases_Success_ReturnsReleases(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "REL-001", "name": "Release 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	releases, meta, err := client.ProjectReleases(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, releases, 1)
	assert.NotNil(t, meta)
}

func TestClient_OpenProjectReleases_Success_ReturnsOpenReleases(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfList:1", "code": "REL-001", "cache_status_type": "OPEN"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	releases, meta, err := client.OpenProjectReleases(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, releases, 1)
	assert.NotNil(t, meta)
}
