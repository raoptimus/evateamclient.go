package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TaskLink_Success_ReturnsTaskLink(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfRelationOption:123",
			"code": "RLO-001",
			"name": "blocks"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.get")
	}

	link, meta, err := client.TaskLink(testCtx, "CmfRelationOption:123", nil)

	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, "CmfRelationOption:123", link.ID)
	assert.Equal(t, "RLO-001", link.Code)
	require.NotNil(t, link.Name)
	assert.Equal(t, "blocks", *link.Name)
	assert.NotNil(t, meta)
}

func TestClient_TaskLink_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "TaskLink not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	link, meta, err := client.TaskLink(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TaskLink not found")
	assert.Nil(t, link)
	assert.Nil(t, meta)
}

func TestClient_TaskLink_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	link, meta, err := client.TaskLink(testCtx, "CmfRelationOption:123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, link)
	assert.Nil(t, meta)
}

func TestClient_TaskLinkQuery_Success_ReturnsTaskLink(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfRelationOption:123",
			"code": "RLO-001",
			"name": "blocks"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.get")
	}

	qb := NewQueryBuilder().From(EntityRelation)
	link, meta, err := client.TaskLinkQuery(testCtx, qb)

	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, "CmfRelationOption:123", link.ID)
	assert.NotNil(t, meta)
}

func TestClient_TaskLinksListQuery_Success_ReturnsTaskLinks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfRelationOption:1", "code": "RLO-001", "name": "blocks"},
			{"id": "CmfRelationOption:2", "code": "RLO-002", "name": "depends on"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.list")
	}

	qb := NewQueryBuilder().From(EntityRelation)
	links, meta, err := client.TaskLinksListQuery(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, links, 2)
	require.NotNil(t, links[0].Name)
	assert.Equal(t, "blocks", *links[0].Name)
	require.NotNil(t, links[1].Name)
	assert.Equal(t, "depends on", *links[1].Name)
	assert.NotNil(t, meta)
}

func TestClient_TaskLinkCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 15
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.count")
	}

	qb := NewQueryBuilder().From(EntityRelation)
	count, err := client.TaskLinkCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 15, count)
}

func TestClient_TaskLinksOutgoing_Success_ReturnsLinks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfRelationOption:1", "code": "RLO-001", "name": "blocks"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.list")
	}

	links, meta, err := client.TaskLinksOutgoing(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, links, 1)
	require.NotNil(t, links[0].Name)
	assert.Equal(t, "blocks", *links[0].Name)
	assert.NotNil(t, meta)
}

func TestClient_TaskLinksIncoming_Success_ReturnsLinks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfRelationOption:2", "code": "RLO-002", "name": "depends on"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.list")
	}

	links, meta, err := client.TaskLinksIncoming(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, links, 1)
	require.NotNil(t, links[0].Name)
	assert.Equal(t, "depends on", *links[0].Name)
	assert.NotNil(t, meta)
}

func TestClient_TaskLinksList_Success_ReturnsLinks(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfRelationOption:1", "code": "RLO-001", "name": "blocks"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.list")
	}

	links, meta, err := client.TaskLinksList(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, links, 1)
	assert.NotNil(t, meta)
}

func TestClient_TaskLinkCreate_Success_ReturnsTaskLink(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfRelationOption:new-123",
			"code": "RLO-100",
			"name": "blocks"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.create")
	}

	link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", "RLO-100")

	require.NoError(t, err)
	assert.Equal(t, "CmfRelationOption:new-123", link.ID)
	require.NotNil(t, link.Name)
	assert.Equal(t, "blocks", *link.Name)
}

func TestClient_TaskLinkDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.delete")
	}

	err := client.TaskLinkDelete(testCtx, "CmfRelationOption:123")

	assert.NoError(t, err)
}
