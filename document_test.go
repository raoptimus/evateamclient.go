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
