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

func TestClient_Document_Success_ReturnsDocument(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfDocument:123",
			"code": "DOC-001",
			"name": "Test Document",
			"text": "Document content"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.get")
	}

	doc, meta, err := client.Document(testCtx, "DOC-001", nil)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "CmfDocument:123", doc.ID)
	assert.Equal(t, "DOC-001", doc.Code)
	assert.Equal(t, "Test Document", doc.Name)
	assert.NotNil(t, meta)
}

func TestClient_Document_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Document not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	doc, meta, err := client.Document(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Document not found")
	assert.Nil(t, doc)
	assert.Nil(t, meta)
}

func TestClient_Document_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	doc, meta, err := client.Document(testCtx, "DOC-001", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, doc)
	assert.Nil(t, meta)
}

func TestClient_DocumentsList_Success_ReturnsDocuments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfDocument:1", "code": "DOC-001", "name": "Doc 1"},
			{"id": "CmfDocument:2", "code": "DOC-002", "name": "Doc 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.list")
	}

	qb := NewQueryBuilder().From(EntityDocument)
	docs, meta, err := client.DocumentsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Equal(t, "DOC-001", docs[0].Code)
	assert.Equal(t, "DOC-002", docs[1].Code)
	assert.NotNil(t, meta)
}

func TestClient_DocumentCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 50
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.count")
	}

	qb := NewQueryBuilder().From(EntityDocument)
	count, err := client.DocumentCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 50, count)
}

func TestClient_ProjectDocuments_Success_ReturnsDocuments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfDocument:1", "code": "DOC-001", "project_id": "CmfProject:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	docs, meta, err := client.ProjectDocuments(testCtx, "CmfProject:123", nil)

	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.NotNil(t, meta)
}

func TestClient_Documents_Success_ReturnsDocuments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfDocument:1", "code": "DOC-001", "name": "Doc 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.list")
	}

	docs, meta, err := client.Documents(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.NotNil(t, meta)
}

func TestClient_DocumentPageTree_Success_ReturnsDocuments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{
				"id": "CmfDocument:aaa",
				"code": "DOC-001",
				"name": "Architecture",
				"parent_id": "CmfProject:123",
				"project_id": "CmfProject:123",
				"orderno": 101500,
				"tree_node_is_branch": true,
				"cache_status_type": "CLOSED"
			},
			{
				"id": "CmfDocument:bbb",
				"code": "DOC-002",
				"name": "Spec Document",
				"parent_id": "CmfDocument:aaa",
				"project_id": "CmfProject:123",
				"orderno": 103375,
				"tree_node_is_branch": false,
				"cache_status_type": "CLOSED"
			}
		],
		"meta": {}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.macros_page_tree_get")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "CmfDocument:root-123")
	}

	docs, err := client.DocumentPageTree(testCtx, "CmfDocument:root-123")

	require.NoError(t, err)
	assert.Len(t, docs, 2)

	assert.Equal(t, "CmfDocument:aaa", docs[0].ID)
	assert.Equal(t, "Architecture", docs[0].Name)
	assert.Equal(t, 101500, docs[0].OrderNo)
	assert.True(t, docs[0].TreeNodeIsBranch)
	assert.Equal(t, "CmfProject:123", docs[0].ParentID)

	assert.Equal(t, "CmfDocument:bbb", docs[1].ID)
	assert.False(t, docs[1].TreeNodeIsBranch)
	assert.Equal(t, "CmfDocument:aaa", docs[1].ParentID)
}

func TestClient_DocumentPageTree_RPCError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Node not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	docs, err := client.DocumentPageTree(testCtx, "CmfDocument:nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Node not found")
	assert.Nil(t, docs)
}

func TestClient_DocumentPageTree_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	docs, err := client.DocumentPageTree(testCtx, "CmfDocument:123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, docs)
}

func TestClient_DocumentCreate_Success_ReturnsDocument(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfDocument:new-123",
			"code": "DOC-100",
			"name": "New Document"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New Document")
	}

	params := DocumentCreateParams{
		Name:      "New Document",
		ProjectID: "CmfProject:123",
	}
	doc, err := client.DocumentCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfDocument:new-123", doc.ID)
	assert.Equal(t, "New Document", doc.Name)
}

// TestClient_DocumentCreate_Success_SendsExactKwargs is the regression test
// for the OAS CmfDocument.create kwargs: {name, parent, tree_parent}, not the
// read/filter field names project_id/parent_id (additionalProperties: false).
func TestClient_DocumentCreate_Success_SendsExactKwargs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.response = mockResponse(http.StatusOK, `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfDocument:new-123", "code": "DOC-100", "name": "New Document"}
	}`)
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, map[string]any{
			"name":        "New Document",
			"parent":      "CmfProject:123",
			"tree_parent": "CmfDocument:root",
		}, parsed.Kwargs)
	}

	params := DocumentCreateParams{
		Name:      "New Document",
		ProjectID: "CmfProject:123",
		ParentID:  "CmfDocument:root",
	}
	doc, err := client.DocumentCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfDocument:new-123", doc.ID)
}

// TestClient_DocumentCreate_WithText_SendsTextDraftAndPublishes covers the
// draft/publish behavior: Text goes into the text_draft kwarg (OAS: no plain
// `text` kwarg on create), and a non-empty Text triggers an automatic
// CmfDocument.do_publish so the created document isn't left invisible.
func TestClient_DocumentCreate_WithText_SendsTextDraftAndPublishes(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:new-123", "code": "DOC-100", "name": "New Document"}
		}`),
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":true}`),
	}

	var createKwargs map[string]any
	var publishMethod string
	var publishArgs []any
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Method string         `json:"method"`
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		switch parsed.Method {
		case "CmfDocument.create":
			createKwargs = parsed.Kwargs
		case "CmfDocument.do_publish":
			publishMethod = parsed.Method
			publishArgs = parsed.Args
		}
		return true
	}

	params := DocumentCreateParams{Name: "New Document", ProjectID: "CmfProject:123", Text: "Document content"}
	doc, err := client.DocumentCreate(testCtx, params)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, map[string]any{
		"name":       "New Document",
		"parent":     "CmfProject:123",
		"text_draft": "Document content",
	}, createKwargs)
	assert.Equal(t, "CmfDocument.do_publish", publishMethod)
	assert.Equal(t, []any{"CmfDocument:new-123"}, publishArgs)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + do_publish")
}

// TestClient_DocumentCreate_WithText_PublishFails_ReturnsDocumentAndWrappedError
// checks that a publish failure isn't silently swallowed (the create would
// otherwise look successful while the text stays an invisible draft), and
// that the created document itself is still returned alongside the error —
// losing it would make a caller retry the create and produce a duplicate
// (SPEC-04 #2).
func TestClient_DocumentCreate_WithText_PublishFails_ReturnsDocumentAndWrappedError(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:new-123", "code": "DOC-100", "name": "New Document"}
		}`),
	}
	mockHTTP.errors = []error{nil, errors.New("publish failed")}

	params := DocumentCreateParams{Name: "New Document", ProjectID: "CmfProject:123", Text: "Document content"}
	doc, err := client.DocumentCreate(testCtx, params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "document CmfDocument:new-123 created, publish failed; do not retry create")
	assert.Contains(t, err.Error(), "publish failed")
	require.NotNil(t, doc)
	assert.Equal(t, "CmfDocument:new-123", doc.ID, "created document must not be lost on publish failure")
}

func TestClient_DocumentCreate_ResultIsIDString_FetchesCreatedDocument(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfDocument:new-123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:new-123", "code": "DOC-100", "name": "New Document"}
		}`),
	}

	params := DocumentCreateParams{Name: "New Document", ProjectID: "CmfProject:123"}
	doc, err := client.DocumentCreate(testCtx, params)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "CmfDocument:new-123", doc.ID)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + follow-up get")
}

// TestClient_DocumentCreate_FollowUpGet_FiltersByIDOrCode covers both forms a
// bare create-result string can take: an ID ("CmfDocument:uuid", which
// carries a ":" class-name prefix) or a code ("DOC-000123", which doesn't).
// Filtering the follow-up .get by the wrong field would find nothing and
// surface as a spurious error, prompting callers to retry and create dupes.
func TestClient_DocumentCreate_FollowUpGet_FiltersByIDOrCode(t *testing.T) {
	tests := []struct {
		name        string
		resultValue string
		wantField   string
	}{
		{"ID form filters by id", "CmfDocument:new-123", DocumentFieldID},
		{"code form filters by code", "DOC-000123", DocumentFieldCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClientWithSequentialMock(t)

			mockHTTP.responses = []*req.Response{
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"`+tt.resultValue+`"}`),
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":{"id":"CmfDocument:new-123","code":"DOC-000123"}}`),
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
				if parsed.Method != "CmfDocument.get" {
					return true
				}
				filter, ok := parsed.Kwargs["filter"].([]any)
				if !assert.True(t, ok, "expected filter kwarg on follow-up .get") {
					return false
				}
				followUpFilter = filter
				return true
			}

			params := DocumentCreateParams{Name: "New Document", ProjectID: "CmfProject:123"}
			doc, err := client.DocumentCreate(testCtx, params)

			require.NoError(t, err)
			require.NotNil(t, doc)
			require.NotEmpty(t, followUpFilter, "follow-up .get should have been called with a filter")
			assert.Equal(t, tt.wantField, followUpFilter[0])
		})
	}
}

// TestClient_DocumentCreate_EmptyResult_ReturnsError guards against a silent
// failure (empty/null/false result) being mistaken for a zero-value document.
func TestClient_DocumentCreate_EmptyResult_ReturnsError(t *testing.T) {
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

			params := DocumentCreateParams{Name: "New Document", ProjectID: "CmfProject:123"}
			doc, err := client.DocumentCreate(testCtx, params)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CmfDocument.create returned empty result")
			assert.Nil(t, doc)
		})
	}
}

// TestClient_DocumentUpdate_Success_SendsIDInArgs is the regression test for
// the EVA update convention: the id belongs in Args, not Kwargs (see TaskUpdate).
func TestClient_DocumentUpdate_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfDocument:123",
			"code": "DOC-001",
			"name": "Updated Document"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.update")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfDocument:123"}, parsed.Args) &&
			assert.Equal(t, map[string]any{"name": "Updated Document"}, parsed.Kwargs)
	}

	updates := map[string]any{"name": "Updated Document"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Document", doc.Name)
}

func TestClient_DocumentUpdate_ResultIsIDString_FetchesUpdatedDocument(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfDocument:123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:123", "code": "DOC-001", "name": "Updated Document"}
		}`),
	}

	updates := map[string]any{"name": "Updated Document"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "Updated Document", doc.Name)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected update + follow-up get")
}

// TestClient_DocumentUpdate_WithText_SendsTextDraftAndPublishes is the
// regression test for SPEC-04 #3: a `text` key in updates has no OAS kwarg on
// CmfDocument.update (additionalProperties: false), so it must be sent as
// text_draft, and — since the caller passing `text` wants a visible result —
// followed by an automatic CmfDocument.do_publish.
func TestClient_DocumentUpdate_WithText_SendsTextDraftAndPublishes(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:123", "code": "DOC-001", "name": "Updated Document"}
		}`),
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":true}`),
	}

	var updateKwargs map[string]any
	var publishMethod string
	var publishArgs []any
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Method string         `json:"method"`
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		switch parsed.Method {
		case "CmfDocument.update":
			updateKwargs = parsed.Kwargs
		case "CmfDocument.do_publish":
			publishMethod = parsed.Method
			publishArgs = parsed.Args
		}
		return true
	}

	updates := map[string]any{"text": "Updated content"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, map[string]any{"text_draft": "Updated content"}, updateKwargs)
	assert.Equal(t, "CmfDocument.do_publish", publishMethod)
	assert.Equal(t, []any{"CmfDocument:123"}, publishArgs)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected update + do_publish")
	assert.Equal(t, map[string]any{"text": "Updated content"}, updates, "input map must not be mutated")
}

// TestClient_DocumentUpdate_WithTextDraft_DoesNotPublish covers the "save a
// draft" path from SPEC-04 #3: passing text_draft directly is an explicit
// choice not to publish, so DocumentUpdate must not call do_publish.
func TestClient_DocumentUpdate_WithTextDraft_DoesNotPublish(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.response = mockResponse(http.StatusOK, `{
		"jsonrpc": "2.2",
		"result": {"id": "CmfDocument:123", "code": "DOC-001", "name": "Updated Document"}
	}`)
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.NotContains(t, string(body), "do_publish")
	}

	updates := map[string]any{"text_draft": "Draft content"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, 1, mockHTTP.calls, "text_draft must not trigger publish")
}

// TestClient_DocumentUpdate_WithText_PublishFails_ReturnsDocumentAndWrappedError
// mirrors the create-side SPEC-04 #2 fix: an update that changed the text but
// failed to publish must still return the updated document, not lose it.
func TestClient_DocumentUpdate_WithText_PublishFails_ReturnsDocumentAndWrappedError(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfDocument:123", "code": "DOC-001", "name": "Updated Document"}
		}`),
	}
	mockHTTP.errors = []error{nil, errors.New("publish failed")}

	updates := map[string]any{"text": "Updated content"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "document CmfDocument:123 updated, publish failed; do not retry update")
	assert.Contains(t, err.Error(), "publish failed")
	require.NotNil(t, doc)
	assert.Equal(t, "CmfDocument:123", doc.ID, "updated document must not be lost on publish failure")
}

func TestClient_DocumentUpdate_EmptyDocID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	doc, err := client.DocumentUpdate(testCtx, "", map[string]any{"name": "x"})

	require.Error(t, err)
	assert.Nil(t, doc)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}

// TestClient_DocumentDelete_Success_SendsIDInArgs is the regression test for
// the EVA delete convention: the id belongs in Args, not Kwargs (see TaskDelete).
func TestClient_DocumentDelete_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.delete")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfDocument:123"}, parsed.Args) &&
			assert.Nil(t, parsed.Kwargs)
	}

	err := client.DocumentDelete(testCtx, "CmfDocument:123")

	assert.NoError(t, err)
}

func TestClient_DocumentDelete_EmptyDocID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	err := client.DocumentDelete(testCtx, "")

	assert.Error(t, err)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}
