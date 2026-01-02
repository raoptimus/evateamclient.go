package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// TimeLog field constants for type-safe queries
const (
	// Core fields
	TimeLogFieldID        = "id"
	TimeLogFieldClassName = "class_name"
	TimeLogFieldCode      = "code"
	TimeLogFieldName      = "name"
	TimeLogFieldTimeSpent = "time_spent"

	// Relations
	TimeLogFieldParent   = "parent"    // nested task object
	TimeLogFieldParentID = "parent_id" // task ID (CmfTask:uuid)

	// System
	TimeLogFieldProjectID     = "project_id"
	TimeLogFieldCmfOwnerID    = "cmf_owner_id" // person who logged time
	TimeLogFieldCmfCreatedAt  = "cmf_created_at"
	TimeLogFieldCmfModifiedAt = "cmf_modified_at"
)

var (
	// DefaultTimeLogFields - standard projection for single time log queries
	DefaultTimeLogFields = []string{
		TimeLogFieldID,
		TimeLogFieldCode,
		TimeLogFieldTimeSpent,
		TimeLogFieldParent,
		TimeLogFieldParentID,
		TimeLogFieldProjectID,
		TimeLogFieldCmfOwnerID,
		TimeLogFieldCmfCreatedAt,
	}

	// DefaultTimeLogListFields - optimized for LIST queries (lighter payload)
	DefaultTimeLogListFields = []string{
		TimeLogFieldID,
		TimeLogFieldCode,
		TimeLogFieldTimeSpent,
		TimeLogFieldParent,
		TimeLogFieldCmfCreatedAt,
	}
)

// TimeLog retrieves a single time log entry by ID (backward compatible)
// Example:
//
//	log, meta, err := client.TimeLog(ctx, "CmfTimeTrackerHistory:uuid", nil)
func (c *Client) TimeLog(
	ctx context.Context,
	timeLogID string,
	fields []string,
) (*models.TimeLog, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldID: timeLogID}).
		Limit(1)

	return c.TimeLogQuery(ctx, qb)
}

// TimeLogQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "time_spent", "author", "description").
//	  From(evateamclient.EntityTimeLog).
//	  Where(sq.Eq{"id": "CmfTimeTrackerHistory:uuid"})
//	log, meta, err := client.TimeLogQuery(ctx, qb)
func (c *Client) TimeLogQuery(ctx context.Context, qb *QueryBuilder) (*models.TimeLog, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTimeLogFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TimeLogResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// TimeLogsList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "time_spent", "author", "description").
//	  From(evateamclient.EntityTimeLog).
//	  Where(sq.Eq{"parent": "CmfTask:uuid"}).
//	  OrderBy("-cmf_created_at").
//	  Offset(0).Limit(100)
//	logs, meta, err := client.TimeLogsList(ctx, qb)
func (c *Client) TimeLogsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.TimeLog, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTimeLogListFields
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

	var resp models.TimeLogListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// TimeLogCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityTimeLog).
//	  Where(sq.Eq{"parent": "CmfTask:uuid"})
//	count, err := client.TimeLogCount(ctx, qb)
func (c *Client) TimeLogCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.count",
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

// TaskTimeLogs retrieves ALL time entries for specific task
// Example:
//
//	logs, meta, err := client.TaskTimeLogs(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskTimeLogs(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TimeLog, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldParentID: taskID}).
		OrderBy("-" + TimeLogFieldCmfCreatedAt)

	return c.TimeLogsList(ctx, qb)
}

// UserTimeLogs retrieves ALL time entries by specific user
// Example:
//
//	logs, meta, err := client.UserTimeLogs(ctx, "CmfPerson:uuid", nil)
func (c *Client) UserTimeLogs(
	ctx context.Context,
	userID string,
	fields []string,
) ([]models.TimeLog, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldCmfOwnerID: userID}).
		OrderBy("-" + TimeLogFieldCmfCreatedAt)

	return c.TimeLogsList(ctx, qb)
}

// UserTaskTimeLogs retrieves time entries for task by specific user
// Example:
//
//	logs, meta, err := client.UserTaskTimeLogs(ctx, "CmfTask:uuid", "CmfPerson:uuid", nil)
func (c *Client) UserTaskTimeLogs(
	ctx context.Context,
	taskID,
	userID string,
	fields []string,
) ([]models.TimeLog, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldParentID: taskID}).
		Where(sq.Eq{TimeLogFieldCmfOwnerID: userID}).
		OrderBy("-" + TimeLogFieldCmfCreatedAt)

	return c.TimeLogsList(ctx, qb)
}

// ProjectTimeLogs retrieves ALL time entries for project tasks (backward compatible)
// Note: This uses dot notation for nested field filtering
// Example:
//
//	logs, meta, err := client.ProjectTimeLogs(ctx, "Project:uuid", nil)
func (c *Client) ProjectTimeLogs(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.TimeLog, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogListFields
	}

	// Using dot notation for nested field filtering
	kwargs := map[string]any{
		"filter":   []any{"parent.project_id", "==", projectID},
		"fields":   fields,
		"order_by": []string{"-" + TimeLogFieldCmfCreatedAt},
	}

	return c.TimeLogs(ctx, kwargs)
}

// Backward compatible methods (using old API)

// TimeLogs retrieves time logs with custom filters (backward compatible, deprecated)
// Recommended: use TimeLogsList with NewQueryBuilder() instead
func (c *Client) TimeLogs(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.TimeLog, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTimeLogListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TimeLogListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// CRUD Operations

// TimeLogCreateParams contains parameters for creating a new time log entry
type TimeLogCreateParams struct {
	ParentID  string `json:"parent_id"`  // task ID (CmfTask:uuid)
	TimeSpent int    `json:"time_spent"` // minutes
}

// TimeLogCreate creates a new time log entry
// Example:
//
//	log, err := client.TimeLogCreate(ctx, TimeLogCreateParams{
//	  ParentID:  "CmfTask:uuid",
//	  TimeSpent: 180, // 3 hours in minutes
//	})
func (c *Client) TimeLogCreate(
	ctx context.Context,
	params TimeLogCreateParams,
) (*models.TimeLog, error) {
	kwargs := map[string]any{
		"parent_id":  params.ParentID,
		"time_spent": params.TimeSpent,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TimeLogResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// TimeLogUpdate updates an existing time log entry
// Example:
//
//	updates := map[string]any{
//	  "time_spent": 240, // 4 hours in minutes
//	}
//	log, err := client.TimeLogUpdate(ctx, "CmfTimeTrackerHistory:uuid", updates)
func (c *Client) TimeLogUpdate(
	ctx context.Context,
	timeLogID string,
	updates map[string]any,
) (*models.TimeLog, error) {
	kwargs := map[string]any{
		"id": timeLogID,
	}
	for k, v := range updates {
		kwargs[k] = v
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.update",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TimeLogResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// TimeLogDelete deletes a time log entry by ID
// Example:
//
//	err := client.TimeLogDelete(ctx, "CmfTimeTrackerHistory:uuid")
func (c *Client) TimeLogDelete(
	ctx context.Context,
	timeLogID string,
) error {
	kwargs := map[string]any{
		"id": timeLogID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}
