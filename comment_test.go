package evateamclient

import (
	"errors"
	"net/http"
	"testing"

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
		return assert.Contains(t, url, "m=Comment.get")
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
		return assert.Contains(t, url, "m=Comment.list")
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
		return assert.Contains(t, url, "m=Comment.count")
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
		return assert.Contains(t, url, "m=Comment.list")
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
		return assert.Contains(t, url, "m=Comment.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New comment")
	}

	comment, err := client.CommentCreate(testCtx, "Task:PROJ-123", "New comment")

	require.NoError(t, err)
	assert.Equal(t, "Comment:new-123", comment.ID)
	assert.Equal(t, "New comment", comment.Text)
}

func TestClient_CommentUpdate_Success_ReturnsUpdatedComment(t *testing.T) {
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
		return assert.Contains(t, url, "m=Comment.update")
	}

	comment, err := client.CommentUpdate(testCtx, "Comment:123", "Updated comment")

	require.NoError(t, err)
	assert.Equal(t, "Updated comment", comment.Text)
}

func TestClient_CommentDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=Comment.delete")
	}

	err := client.CommentDelete(testCtx, "Comment:123")

	assert.NoError(t, err)
}
