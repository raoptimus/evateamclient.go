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

// TestClient_TaskLinkCreate_Success_SendsExactKwargs is the regression test for
// the original bug: CmfRelationOption.create only accepts out_link/in_link/
// relation_type kwargs (additionalProperties: false in the OpenAPI spec). A
// stray "id" key (the old, wrong kwarg) makes this test fail.
func TestClient_TaskLinkCreate_Success_SendsExactKwargs(t *testing.T) {
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
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, map[string]any{
			"out_link":      "CmfTask:source",
			"in_link":       "CmfTask:target",
			"relation_type": RelationTypeLink,
		}, parsed.Kwargs)
	}

	link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", RelationTypeLink)

	require.NoError(t, err)
	assert.Equal(t, "CmfRelationOption:new-123", link.ID)
	require.NotNil(t, link.Name)
	assert.Equal(t, "blocks", *link.Name)
}

func TestClient_TaskLinkCreate_ResultIsIDString_FetchesCreatedLink(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfRelationOption:new-123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {
				"id": "CmfRelationOption:new-123",
				"code": "RLO-100",
				"name": "blocks",
				"relation_type": "system.link"
			}
		}`),
	}

	link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", RelationTypeLink)

	require.NoError(t, err)
	require.NotNil(t, link)
	assert.Equal(t, "CmfRelationOption:new-123", link.ID)
	assert.Equal(t, "system.link", string(link.RelationType))
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + follow-up get")
}

// TestClient_TaskLinkCreate_FollowUpGetReturnsEmptyLink_ReturnsError guards
// against the follow-up .get silently succeeding with a zero-value TaskLink
// (empty ID) — that is exactly the "silent empty object" the create fix was
// meant to eliminate.
func TestClient_TaskLinkCreate_FollowUpGetReturnsEmptyLink_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfRelationOption:new-123"}`),
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":{}}`),
	}

	link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", RelationTypeLink)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "CmfRelationOption.create returned empty result")
	assert.Nil(t, link)
}

// TestClient_TaskLinkCreate_FollowUpGet_FiltersByIDOrCode covers both forms a
// bare create-result string can take: an ID ("CmfRelationOption:uuid", which
// carries a ":" class-name prefix) or a code ("RLO-000123", which doesn't).
// Filtering the follow-up .get by the wrong field would find nothing and
// surface as a spurious error, prompting callers to retry and create dupes.
func TestClient_TaskLinkCreate_FollowUpGet_FiltersByIDOrCode(t *testing.T) {
	tests := []struct {
		name        string
		resultValue string
		wantField   string
	}{
		{"ID form filters by id", "CmfRelationOption:new-123", TaskLinkFieldID},
		{"code form filters by code", "RLO-000123", TaskLinkFieldCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClientWithSequentialMock(t)

			mockHTTP.responses = []*req.Response{
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"`+tt.resultValue+`"}`),
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":{"id":"CmfRelationOption:new-123","code":"RLO-000123"}}`),
			}

			var followUpFilter []any
			mockHTTP.bodyCheck = func(body []byte) bool {
				var parsed struct {
					Method string         `json:"method"`
					Kwargs map[string]any `json:"kwargs"`
				}
				if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
					return false
				}
				if parsed.Method != "CmfRelationOption.get" {
					return true
				}
				filter, ok := parsed.Kwargs["filter"].([]any)
				if !assert.True(t, ok, "expected filter kwarg on follow-up .get") {
					return false
				}
				followUpFilter = filter
				return true
			}

			link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", RelationTypeLink)

			require.NoError(t, err)
			require.NotNil(t, link)
			require.NotEmpty(t, followUpFilter, "follow-up .get should have been called with a filter")
			assert.Equal(t, tt.wantField, followUpFilter[0])
		})
	}
}

func TestClient_TaskLinkCreate_EmptyResult_ReturnsError(t *testing.T) {
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

			link, err := client.TaskLinkCreate(testCtx, "CmfTask:source", "CmfTask:target", RelationTypeLink)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CmfRelationOption.create returned empty result")
			assert.Nil(t, link)
		})
	}
}

func TestClient_TaskLinkCreate_EmptyArgument_ReturnsErrorWithoutRequest(t *testing.T) {
	tests := []struct {
		name                          string
		outLink, inLink, relationType string
	}{
		{"empty outLink", "", "CmfTask:target", RelationTypeLink},
		{"empty inLink", "CmfTask:source", "", RelationTypeLink},
		{"empty relationType", "CmfTask:source", "CmfTask:target", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClient(t)
			mockHTTP.err = errors.New("must not be called")

			link, err := client.TaskLinkCreate(testCtx, tt.outLink, tt.inLink, tt.relationType)

			require.Error(t, err)
			assert.Nil(t, link)
			assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
		})
	}
}

// TestClient_TaskLinkDelete_Success_SendsIDInArgs is the regression test for
// the EVA delete convention: the id belongs in Args, not Kwargs (see TaskDelete).
func TestClient_TaskLinkDelete_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfRelationOption.delete")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfRelationOption:123"}, parsed.Args) &&
			assert.Nil(t, parsed.Kwargs)
	}

	err := client.TaskLinkDelete(testCtx, "CmfRelationOption:123")

	assert.NoError(t, err)
}

// TestClient_TaskLinkDelete_NonBooleanResult_Succeeds pins the TaskDelete
// convention (task.go, Result any): CmfRelationOption.delete has no fixed
// response shape in the OAS, so a non-bool result (an ID echo, or an empty
// object) must not be reported as an error on an otherwise successful delete.
func TestClient_TaskLinkDelete_NonBooleanResult_Succeeds(t *testing.T) {
	tests := []struct {
		name   string
		result string
	}{
		{"id string", `"CmfRelationOption:1"`},
		{"empty object", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClient(t)
			mockHTTP.response = mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":`+tt.result+`}`)

			err := client.TaskLinkDelete(testCtx, "CmfRelationOption:123")

			assert.NoError(t, err)
		})
	}
}

func TestClient_TaskLinkDelete_EmptyLinkID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	err := client.TaskLinkDelete(testCtx, "")

	assert.Error(t, err)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}

func TestClient_TaskLink_InLinkOutLink_StringForm_ParsesAsID(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfRelationOption:123",
			"code": "RLO-001",
			"relation_type": "system.link",
			"in_link": "CmfTask:target",
			"out_link": "CmfTask:source"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	link, _, err := client.TaskLink(testCtx, "CmfRelationOption:123", nil)

	require.NoError(t, err)
	assert.Equal(t, "system.link", string(link.RelationType))
	require.NotNil(t, link.InLink)
	assert.Equal(t, "CmfTask:target", link.InLink.ID)
	require.NotNil(t, link.OutLink)
	assert.Equal(t, "CmfTask:source", link.OutLink.ID)
}

func TestClient_TaskLink_InLinkOutLink_ObjectForm_ParsesFields(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfRelationOption:123",
			"code": "RLO-001",
			"relation_type": "system.link",
			"in_link": {"id": "CmfTask:target", "code": "TSK-000002", "name": "Target task"},
			"out_link": {"id": "CmfTask:source", "code": "TSK-000001", "name": "Source task"}
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	link, _, err := client.TaskLink(testCtx, "CmfRelationOption:123", AllBasicAndRelationFields)

	require.NoError(t, err)
	require.NotNil(t, link.InLink)
	assert.Equal(t, "CmfTask:target", link.InLink.ID)
	assert.Equal(t, "TSK-000002", link.InLink.Code)
	assert.Equal(t, "Target task", link.InLink.Name)
	require.NotNil(t, link.OutLink)
	assert.Equal(t, "CmfTask:source", link.OutLink.ID)
	assert.Equal(t, "TSK-000001", link.OutLink.Code)
	assert.Equal(t, "Source task", link.OutLink.Name)
}
