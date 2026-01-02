package evateamclient

import (
	"errors"
	"net/http"
	"testing"

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

func TestClient_DocumentUpdate_Success_ReturnsUpdatedDocument(t *testing.T) {
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

	updates := map[string]any{"name": "Updated Document"}
	doc, err := client.DocumentUpdate(testCtx, "CmfDocument:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Document", doc.Name)
}

func TestClient_DocumentDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfDocument.delete")
	}

	err := client.DocumentDelete(testCtx, "CmfDocument:123")

	assert.NoError(t, err)
}
