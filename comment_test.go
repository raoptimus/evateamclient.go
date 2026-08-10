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

func TestClient_Comment_Success_ReturnsComment(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "Comment:123",
			"text": "This is a comment",
			"task_id": "Task:PROJ-001"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.get")
	}

	comment, meta, err := client.Comment(testCtx, "Comment:123", nil)

	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, "Comment:123", comment.ID)
	assert.Equal(t, "This is a comment", comment.Text)
	assert.NotNil(t, meta)
}

func TestClient_Comment_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Comment not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	comment, meta, err := client.Comment(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Comment not found")
	assert.Nil(t, comment)
	assert.Nil(t, meta)
}

func TestClient_Comment_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	comment, meta, err := client.Comment(testCtx, "Comment:123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, comment)
	assert.Nil(t, meta)
}

func TestClient_CommentsList_Success_ReturnsComments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "Comment:1", "text": "Comment 1"},
			{"id": "Comment:2", "text": "Comment 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.list")
	}

	qb := NewQueryBuilder().From(EntityComment)
	comments, meta, err := client.CommentsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "Comment 1", comments[0].Text)
	assert.Equal(t, "Comment 2", comments[1].Text)
	assert.NotNil(t, meta)
}

func TestClient_CommentCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 35
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.count")
	}

	qb := NewQueryBuilder().From(EntityComment)
	count, err := client.CommentCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 35, count)
}

func TestClient_TaskComments_Success_ReturnsComments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "Comment:1", "text": "Comment on task", "task_id": "Task:PROJ-123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	comments, meta, err := client.TaskComments(testCtx, "PROJ-123", nil)

	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.NotNil(t, meta)
}

func TestClient_TaskCommentsByID_Success_ReturnsComments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "Comment:1", "text": "Comment on task", "task_id": "CmfTask:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	comments, meta, err := client.TaskCommentsByID(testCtx, "CmfTask:123", nil)

	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.NotNil(t, meta)
}

func TestClient_UserComments_Success_ReturnsComments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "Comment:1", "text": "User comment", "cmf_author_id": "CmfPerson:123"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	comments, meta, err := client.UserComments(testCtx, "CmfPerson:123", nil)

	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.NotNil(t, meta)
}

func TestClient_Comments_Success_ReturnsComments(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "Comment:1", "text": "Comment 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.list")
	}

	comments, meta, err := client.Comments(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.NotNil(t, meta)
}

func TestClient_CommentCreate_Success_ReturnsComment(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "Comment:new-123",
			"text": "New comment",
			"task_id": "Task:PROJ-123"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New comment")
	}

	comment, err := client.CommentCreate(testCtx, "Task:PROJ-123", "New comment")

	require.NoError(t, err)
	assert.Equal(t, "Comment:new-123", comment.ID)
	assert.Equal(t, "New comment", comment.Text)
}

func TestClient_CommentCreate_ResultIsIDString_FetchesCreatedComment(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"Comment:new-123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "Comment:new-123", "text": "New comment"}
		}`),
	}

	comment, err := client.CommentCreate(testCtx, "Task:PROJ-123", "New comment")

	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, "Comment:new-123", comment.ID)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + follow-up get")
}

// TestClient_CommentCreate_EmptyResult_ReturnsError guards against a silent
// failure (empty/null/false result) being mistaken for a zero-value comment.
func TestClient_CommentCreate_EmptyResult_ReturnsError(t *testing.T) {
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

			comment, err := client.CommentCreate(testCtx, "Task:PROJ-123", "New comment")

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CmfComment.create returned empty result")
			assert.Nil(t, comment)
		})
	}
}

// TestClient_CommentUpdate_Success_SendsIDInArgs is the regression test for
// the EVA update convention: the id belongs in Args, not Kwargs (see TaskUpdate).
func TestClient_CommentUpdate_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "Comment:123",
			"text": "Updated comment"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.update")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"Comment:123"}, parsed.Args) &&
			assert.Equal(t, map[string]any{"text": "Updated comment"}, parsed.Kwargs)
	}

	comment, err := client.CommentUpdate(testCtx, "Comment:123", "Updated comment")

	require.NoError(t, err)
	assert.Equal(t, "Updated comment", comment.Text)
}

func TestClient_CommentUpdate_ResultIsIDString_FetchesUpdatedComment(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"Comment:123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "Comment:123", "text": "Updated comment"}
		}`),
	}

	comment, err := client.CommentUpdate(testCtx, "Comment:123", "Updated comment")

	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, "Updated comment", comment.Text)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected update + follow-up get")
}

func TestClient_CommentUpdate_EmptyCommentID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	comment, err := client.CommentUpdate(testCtx, "", "text")

	require.Error(t, err)
	assert.Nil(t, comment)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}

// TestClient_CommentDelete_Success_SendsIDInArgs is the regression test for
// the EVA delete convention: the id belongs in Args, not Kwargs (see TaskDelete).
func TestClient_CommentDelete_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfComment.delete")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"Comment:123"}, parsed.Args) &&
			assert.Nil(t, parsed.Kwargs)
	}

	err := client.CommentDelete(testCtx, "Comment:123")

	assert.NoError(t, err)
}

func TestClient_CommentDelete_EmptyCommentID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	err := client.CommentDelete(testCtx, "")

	assert.Error(t, err)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}
