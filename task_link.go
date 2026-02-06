package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go/models"
)

// TaskLink field constants for type-safe queries
const (
	// Core fields
	TaskLinkFieldID        = "id"
	TaskLinkFieldClassName = "class_name"
	TaskLinkFieldCode      = "code"
	TaskLinkFieldName      = "name"
	TaskLinkFieldInLink    = "in_link"  // filter field: incoming links to task
	TaskLinkFieldOutLink   = "out_link" // filter field: outgoing links from task

	// System
	TaskLinkFieldCmfCreatedAt  = "cmf_created_at"
	TaskLinkFieldCmfModifiedAt = "cmf_modified_at"
	TaskLinkFieldCmfOwnerID    = "cmf_owner_id"
)

var (
	// DefaultTaskLinkFields - standard projection for task link queries
	DefaultTaskLinkFields = []string{
		TaskLinkFieldID,
		TaskLinkFieldClassName,
		TaskLinkFieldCode,
		TaskLinkFieldName,
		TaskLinkFieldCmfCreatedAt,
		TaskLinkFieldCmfOwnerID,
	}

	// DefaultTaskLinkListFields - optimized for LIST queries
	DefaultTaskLinkListFields = []string{
		TaskLinkFieldID,
		TaskLinkFieldCode,
		TaskLinkFieldName,
	}
)

// TaskLink retrieves a single task link by ID
// Example:
//
//	link, meta, err := client.TaskLink(ctx, "CmfTaskLink:uuid", nil)
func (c *Client) TaskLink(
	ctx context.Context,
	linkID string,
	fields []string,
) (*models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldID: linkID}).
		Limit(1)

	return c.TaskLinkQuery(ctx, qb)
}

// TaskLinkQuery executes query using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{"id": "CmfRelationOption:uuid"})
//	link, meta, err := client.TaskLinkQuery(ctx, qb)
func (c *Client) TaskLinkQuery(ctx context.Context, qb *QueryBuilder) (*models.TaskLink, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// TaskLinksListQuery retrieves list using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{evateamclient.TaskLinkFieldOutLink: "CmfTask:uuid"}).
//	  Limit(100)
//	links, meta, err := client.TaskLinksListQuery(ctx, qb)
func (c *Client) TaskLinksListQuery(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.TaskLink, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkListFields
	}

	method, err := qb.ToMethod(false)
	if err != nil {
		return nil, nil, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// TaskLinkCount counts task links using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{evateamclient.TaskLinkFieldOutLink: "CmfTask:uuid"})
//	count, err := client.TaskLinkCount(ctx, qb)
func (c *Client) TaskLinkCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.count",
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

// TaskLinks retrieves ALL task relationships (both directions)
// Makes two API calls (outgoing + incoming) and merges results
// Example:
//
//	links, meta, err := client.TaskLinks(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinks(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	// Get outgoing links (task is source)
	outgoing, _, err := c.TaskLinksOutgoing(ctx, taskID, fields)
	if err != nil {
		return nil, nil, err
	}

	// Get incoming links (task is target)
	incoming, meta, err := c.TaskLinksIncoming(ctx, taskID, fields)
	if err != nil {
		return nil, nil, err
	}

	// Merge results, avoiding duplicates by ID
	seen := make(map[string]bool)
	var result []models.TaskLink

	for _, link := range outgoing {
		if !seen[link.ID] {
			seen[link.ID] = true
			result = append(result, link)
		}
	}
	for _, link := range incoming {
		if !seen[link.ID] {
			seen[link.ID] = true
			result = append(result, link)
		}
	}

	return result, meta, nil
}

// TaskLinksOutgoing retrieves links where task is source (outgoing)
// Example:
//
//	links, meta, err := client.TaskLinksOutgoing(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinksOutgoing(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldOutLink: taskID})

	return c.TaskLinksListQuery(ctx, qb)
}

// TaskLinksIncoming retrieves links where task is target (incoming)
// Example:
//
//	links, meta, err := client.TaskLinksIncoming(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinksIncoming(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldInLink: taskID})

	return c.TaskLinksListQuery(ctx, qb)
}

// TaskLinkCreate creates a new task link
// Example:
//
//	link, err := client.TaskLinkCreate(ctx, "CmfTask:uuid1", "CmfTask:uuid2", "RLO-000001")
func (c *Client) TaskLinkCreate(
	ctx context.Context,
	sourceTaskID, targetTaskID, relationOptionID string,
) (*models.TaskLink, error) {
	kwargs := map[string]any{
		TaskLinkFieldOutLink: sourceTaskID,
		TaskLinkFieldInLink:  targetTaskID,
		"id":                 relationOptionID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// TaskLinkDelete deletes a task link by ID
// Example:
//
//	err := client.TaskLinkDelete(ctx, "RLO-000001")
func (c *Client) TaskLinkDelete(
	ctx context.Context,
	linkID string,
) error {
	kwargs := map[string]any{
		"id": linkID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// Backward compatible methods (using old API)

// TaskLinksList retrieves task links with custom filters (backward compatible, deprecated)
// Recommended: use TaskLinksListQuery with NewQueryBuilder() instead
func (c *Client) TaskLinksList(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.TaskLink, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
