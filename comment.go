package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Comment field constants for type-safe queries
const (
	// Core fields
	CommentFieldID        = "id"
	CommentFieldClassName = "class_name"
	CommentFieldText      = "text"
	CommentFieldLogLevel  = "log_level"

	// Relations
	CommentFieldTaskID   = "task_id" // parent task
	CommentFieldAuthorID = "cmf_author_id"

	// System
	CommentFieldCmfCreatedAt  = "cmf_created_at"
	CommentFieldCmfModifiedAt = "cmf_modified_at"
	CommentFieldCmfOwnerID    = "cmf_owner_id"
)

var (
	// DefaultCommentFields - standard projection for single comment queries
	DefaultCommentFields = []string{
		CommentFieldID,
		CommentFieldClassName,
		CommentFieldText,
		CommentFieldAuthorID,
		CommentFieldTaskID,
		CommentFieldCmfCreatedAt,
		CommentFieldLogLevel,
	}

	// DefaultCommentListFields - optimized for LIST queries (lighter payload)
	DefaultCommentListFields = []string{
		CommentFieldID,
		CommentFieldText,
		CommentFieldAuthorID,
		CommentFieldCmfCreatedAt,
	}
)

// Comment retrieves a single comment by ID
// Example:
//
//	comment, meta, err := client.Comment(ctx, "Comment:uuid", nil)
func (c *Client) Comment(
	ctx context.Context,
	commentID string,
	fields []string,
) (*models.Comment, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldID: commentID}).
		Limit(1)

	return c.CommentQuery(ctx, qb)
}

// CommentQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "text", "cmf_author_id").
//	  From(evateamclient.EntityComment).
//	  Where(sq.Eq{"id": "Comment:uuid"})
//	comment, meta, err := client.CommentQuery(ctx, qb)
func (c *Client) CommentQuery(ctx context.Context, qb *QueryBuilder) (*models.Comment, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultCommentFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// CommentsList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "text", "cmf_author_id", "cmf_created_at").
//	  From(evateamclient.EntityComment).
//	  Where(sq.Eq{"task_id": "Task:PROJ-123"}).
//	  OrderBy("-cmf_created_at").
//	  Offset(0).Limit(100)
//	comments, meta, err := client.CommentsList(ctx, qb)
func (c *Client) CommentsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Comment, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultCommentListFields
	}

	method, err := qb.ToMethod()
	if err != nil {
		return nil, nil, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// CommentCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityComment).
//	  Where(sq.Eq{"task_id": "Task:PROJ-123"})
//	count, err := client.CommentCount(ctx, qb)
func (c *Client) CommentCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  int    `json:"result"`
	}

	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// TaskComments retrieves ALL comments for task (backward compatible)
// Example:
//
//	comments, meta, err := client.TaskComments(ctx, "PROJ-123", nil)
func (c *Client) TaskComments(
	ctx context.Context,
	taskCode string,
	fields []string,
) ([]models.Comment, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldTaskID: "Task:" + taskCode}).
		OrderBy("-" + CommentFieldCmfCreatedAt)

	return c.CommentsList(ctx, qb)
}

// TaskCommentsByID retrieves ALL comments for task by task ID
// Example:
//
//	comments, meta, err := client.TaskCommentsByID(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskCommentsByID(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.Comment, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldTaskID: taskID}).
		OrderBy("-" + CommentFieldCmfCreatedAt)

	return c.CommentsList(ctx, qb)
}

// UserComments retrieves ALL comments by specific user
// Example:
//
//	comments, meta, err := client.UserComments(ctx, "Person:uuid", nil)
func (c *Client) UserComments(
	ctx context.Context,
	userID string,
	fields []string,
) ([]models.Comment, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldAuthorID: userID}).
		OrderBy("-" + CommentFieldCmfCreatedAt)

	return c.CommentsList(ctx, qb)
}

// Backward compatible methods (using old API)

// Comments retrieves comments with custom filters (backward compatible, deprecated)
// Recommended: use CommentsList with NewQueryBuilder() instead
func (c *Client) Comments(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Comment, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultCommentListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// CRUD Operations

// CommentCreate creates a new comment on a task
// Example:
//
//	comment, err := client.CommentCreate(ctx, "Task:PROJ-123", "This is a comment")
func (c *Client) CommentCreate(
	ctx context.Context,
	taskID string,
	text string,
) (*models.Comment, error) {
	kwargs := map[string]any{
		"task_id": taskID,
		"text":    text,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// CommentUpdate updates an existing comment
// Example:
//
//	comment, err := client.CommentUpdate(ctx, "Comment:uuid", "Updated text")
func (c *Client) CommentUpdate(
	ctx context.Context,
	commentID string,
	text string,
) (*models.Comment, error) {
	kwargs := map[string]any{
		"id":   commentID,
		"text": text,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.update",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// CommentDelete deletes a comment by ID
// Example:
//
//	err := client.CommentDelete(ctx, "Comment:uuid")
func (c *Client) CommentDelete(
	ctx context.Context,
	commentID string,
) error {
	kwargs := map[string]any{
		"id": commentID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}
