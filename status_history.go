package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// StatusHistory field constants for type-safe queries
const (
	// Core fields
	StatusHistoryFieldID        = "id"
	StatusHistoryFieldClassName = "class_name"
	StatusHistoryFieldCode      = "code"
	StatusHistoryFieldName      = "name"

	// Relations
	StatusHistoryFieldParentID    = "parent_id"     // entity that changed status
	StatusHistoryFieldProjectID   = "project_id"    // project context
	StatusHistoryFieldOldStatus   = "old_status"    // previous status value
	StatusHistoryFieldNewStatus   = "new_status"    // new status value
	StatusHistoryFieldOldStatusID = "old_status_id" // previous status ID
	StatusHistoryFieldNewStatusID = "new_status_id" // new status ID

	// System
	StatusHistoryFieldCmfOwnerID    = "cmf_owner_id"
	StatusHistoryFieldCmfCreatedAt  = "cmf_created_at"
	StatusHistoryFieldCmfModifiedAt = "cmf_modified_at"
)

var (
	// DefaultStatusHistoryFields - standard projection for single status history queries
	DefaultStatusHistoryFields = []string{
		StatusHistoryFieldID,
		StatusHistoryFieldCode,
		StatusHistoryFieldParentID,
		StatusHistoryFieldOldStatus,
		StatusHistoryFieldNewStatus,
		StatusHistoryFieldCmfOwnerID,
		StatusHistoryFieldCmfCreatedAt,
	}

	// DefaultStatusHistoryListFields - optimized for LIST queries (lighter payload)
	DefaultStatusHistoryListFields = []string{
		StatusHistoryFieldID,
		StatusHistoryFieldCode,
		StatusHistoryFieldParentID,
		StatusHistoryFieldOldStatus,
		StatusHistoryFieldNewStatus,
		StatusHistoryFieldCmfCreatedAt,
	}
)

// StatusHistory retrieves a single status history entry by ID
// Example:
//
//	history, meta, err := client.StatusHistory(ctx, "CmfStatusHistory:uuid", nil)
func (c *Client) StatusHistory(
	ctx context.Context,
	statusHistoryID string,
	fields []string,
) (*models.StatusHistory, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldID: statusHistoryID}).
		Limit(1)

	return c.StatusHistoryQuery(ctx, qb)
}

// StatusHistoryQuery executes query using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "old_status", "new_status", "cmf_created_at").
//	  From(evateamclient.EntityStatusHistory).
//	  Where(sq.Eq{"id": "CmfStatusHistory:uuid"})
//	history, meta, err := client.StatusHistoryQuery(ctx, qb)
func (c *Client) StatusHistoryQuery(ctx context.Context, qb *QueryBuilder) (*models.StatusHistory, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultStatusHistoryFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfStatusHistory.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.StatusHistoryResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// StatusHistoryList retrieves list using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "old_status", "new_status", "cmf_created_at").
//	  From(evateamclient.EntityStatusHistory).
//	  Where(sq.Eq{"parent_id": "CmfTask:uuid"}).
//	  OrderBy("-cmf_created_at").
//	  Offset(0).Limit(100)
//	histories, meta, err := client.StatusHistoryList(ctx, qb)
func (c *Client) StatusHistoryList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.StatusHistory, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultStatusHistoryListFields
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

	var resp models.StatusHistoryListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// StatusHistoryCount counts status history entries using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityStatusHistory).
//	  Where(sq.Eq{"parent_id": "CmfTask:uuid"})
//	count, err := client.StatusHistoryCount(ctx, qb)
func (c *Client) StatusHistoryCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfStatusHistory.count",
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

// TaskStatusHistory retrieves ALL status changes for specific task
// Example:
//
//	histories, meta, err := client.TaskStatusHistory(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskStatusHistory(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.StatusHistory, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldParentID: taskID}).
		OrderBy("-" + StatusHistoryFieldCmfCreatedAt)

	return c.StatusHistoryList(ctx, qb)
}

// ProjectStatusHistory retrieves ALL status changes for project entities
// Example:
//
//	histories, meta, err := client.ProjectStatusHistory(ctx, "CmfProject:uuid", nil)
func (c *Client) ProjectStatusHistory(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.StatusHistory, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldProjectID: projectID}).
		OrderBy("-" + StatusHistoryFieldCmfCreatedAt)

	return c.StatusHistoryList(ctx, qb)
}

// Backward compatible methods (using old API)

// StatusHistories retrieves status histories with custom filters (backward compatible)
// Recommended: use StatusHistoryList with NewQueryBuilder() instead
func (c *Client) StatusHistories(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.StatusHistory, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultStatusHistoryListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfStatusHistory.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.StatusHistoryListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
