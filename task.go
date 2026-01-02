package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Task field constants for type-safe queries
const (
	// Core fields
	TaskFieldID              = "id"
	TaskFieldClassName       = "class_name"
	TaskFieldCode            = "code"
	TaskFieldName            = "name"
	TaskFieldText            = "text"
	TaskFieldProjectID       = "project_id"
	TaskFieldParentID        = "parent_id"
	TaskFieldParentTask      = "parent_task"
	TaskFieldCacheStatusType = "cache_status_type"
	TaskFieldPriority        = "priority"
	TaskFieldDeadline        = "deadline"
	TaskFieldMark            = "mark"
	TaskFieldAlarmDate       = "alarm_date"

	// Story Points
	TaskFieldAgileStoryPoints = "agile_story_points"

	// Relations (single)
	TaskFieldResponsible   = "responsible"
	TaskFieldResponsibleID = "responsible_id"
	TaskFieldWaitingFor    = "waiting_for"
	TaskFieldEpic          = "epic"
	TaskFieldEpicID        = "epic_id"
	TaskFieldLogicType     = "logic_type"
	TaskFieldLogicTypeID   = "logic_type_id"
	TaskFieldCmfOwnerID    = "cmf_owner_id"
	TaskFieldWorkflowID    = "workflow_id"

	// Relations (arrays)
	TaskFieldLists       = "lists" // sprints
	TaskFieldFixVersions = "fix_versions"
	TaskFieldTags        = "tags"
	TaskFieldExecutors   = "executors"
	TaskFieldSpectators  = "spectators"
	TaskFieldComponents  = "components"

	// Status tracking
	TaskFieldStatusModifiedAt      = "status_modified_at"
	TaskFieldStatusInProgressStart = "status_in_progress_start"
	TaskFieldStatusInProgressEnd   = "status_in_progress_end"
	TaskFieldStatusReviewAt        = "status_review_at"
	TaskFieldStatusClosedAt        = "status_closed_at"

	// Planning dates
	TaskFieldPlanStartDate  = "plan_start_date"
	TaskFieldPlanEndDate    = "plan_end_date"
	TaskFieldPeriodInterval = "period_interval"
	TaskFieldPeriodNextDate = "period_next_date"

	// Flags
	TaskFieldApproved  = "approved"
	TaskFieldIsPublic  = "is_public"
	TaskFieldNoControl = "no_control"
	TaskFieldIsFlagged = "is_flagged"

	// System
	TaskFieldCmfCreatedAt         = "cmf_created_at"
	TaskFieldCmfModifiedAt        = "cmf_modified_at"
	TaskFieldCmfViewedAt          = "cmf_viewed_at"
	TaskFieldCmfDeleted           = "cmf_deleted"
	TaskFieldCmfVersion           = "cmf_version"
	TaskFieldCmfLockedAt          = "cmf_locked_at"
	TaskFieldCacheChildTasksCount = "cache_child_tasks_count"
	TaskFieldExtID                = "ext_id"
	TaskFieldArchiveDate          = "archiveddate"
	TaskFieldResultText           = "result_text"
)

var (
	// DefaultTaskFields - standard projection for single task queries
	DefaultTaskFields = []string{
		TaskFieldID,
		TaskFieldCode,
		TaskFieldName,
		TaskFieldText,
		TaskFieldProjectID,
		TaskFieldLists,
		TaskFieldCmfOwnerID,
		TaskFieldResponsible,
		TaskFieldCacheStatusType,
		TaskFieldPriority,
		TaskFieldDeadline,
		TaskFieldEpic,
		TaskFieldTags,
		TaskFieldExecutors,
		TaskFieldWaitingFor,
		TaskFieldParentID,
		TaskFieldFixVersions,
		TaskFieldAgileStoryPoints,
		TaskFieldComponents,
		TaskFieldLogicType,
	}

	// DefaultTaskListFields - optimized for LIST queries (lighter payload)
	DefaultTaskListFields = []string{
		TaskFieldID,
		TaskFieldCode,
		TaskFieldName,
		TaskFieldProjectID,
		TaskFieldCacheStatusType,
		TaskFieldPriority,
		TaskFieldDeadline,
		TaskFieldResponsibleID,
		TaskFieldEpicID,
		TaskFieldAgileStoryPoints,
	}
)

// Task retrieves a single task by code (backward compatible)
// Example:
//
//	task, meta, err := client.Task(ctx, "PROJ-123", nil)
func (c *Client) Task(
	ctx context.Context,
	taskCode string,
	fields []string,
) (*models.Task, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldCode: taskCode}).
		Limit(1)

	return c.TaskQuery(ctx, qb)
}

// TaskQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name", "responsible", "executors").
//	  From(evateamclient.EntityTask).
//	  Where(sq.Eq{"code": "PROJ-123"})
//	task, meta, err := client.TaskQuery(ctx, qb)
func (c *Client) TaskQuery(ctx context.Context, qb *QueryBuilder) (*models.Task, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// TasksList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name", "priority", "cache_status_type").
//	  From(evateamclient.EntityTask).
//	  Where(sq.Eq{"project_id": "Project:uuid"}).
//	  Where(sq.Eq{"cache_status_type": evateamclient.StatusTypeOpen}).
//	  OrderBy("-priority", "name").
//	  Offset(0).Limit(100)
//	tasks, meta, err := client.TasksList(ctx, qb)
func (c *Client) TasksList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Task, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskListFields
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

	var resp models.TaskListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// TaskCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityTask).
//	  Where(sq.Eq{"project_id": "Project:uuid"}).
//	  Where(sq.Eq{"cache_status_type": evateamclient.StatusTypeOpen})
//	count, err := client.TaskCount(ctx, qb)
func (c *Client) TaskCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.count",
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

// ProjectTasks retrieves ALL tasks for project (backward compatible)
// Example:
//
//	tasks, meta, err := client.ProjectTasks(ctx, "Project:uuid", nil)
func (c *Client) ProjectTasks(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldParentID: projectID})

	return c.TasksList(ctx, qb)
}

// SprintTasks retrieves ALL tasks for sprint (backward compatible)
// Example:
//
//	tasks, meta, err := client.SprintTasks(ctx, "SPRINT-CODE", nil)
func (c *Client) SprintTasks(
	ctx context.Context,
	sprintCode string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskListFields
	}

	// Note: "contains" operator is EVA-specific, not standard SQL
	// Using raw kwargs for this special case
	kwargs := map[string]any{
		"filter": []any{TaskFieldLists, "contains", sprintCode},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// PersonTasks retrieves ALL tasks where user is responsible
// Note: For OR logic (responsible OR executor), use PersonTasksAll
// Example:
//
//	tasks, meta, err := client.PersonTasks(ctx, "Person:uuid", nil)
func (c *Client) PersonTasks(
	ctx context.Context,
	userID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldResponsible: userID})

	return c.TasksList(ctx, qb)
}

// PersonTasksAsExecutor retrieves tasks where user is in executors list
// Example:
//
//	tasks, meta, err := client.PersonTasksAsExecutor(ctx, "Person:uuid", nil)
func (c *Client) PersonTasksAsExecutor(
	ctx context.Context,
	userID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskListFields
	}

	// Note: "contains" operator is EVA-specific, not standard SQL
	kwargs := map[string]any{
		"filter": []any{TaskFieldExecutors, "contains", userID},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// PersonProjectTasks retrieves user's tasks as responsible in specific project
// Example:
//
//	tasks, meta, err := client.PersonProjectTasks(ctx, "Project:uuid", "Person:uuid", nil)
func (c *Client) PersonProjectTasks(
	ctx context.Context,
	projectID,
	userID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Where(sq.Eq{TaskFieldResponsible: userID})

	return c.TasksList(ctx, qb)
}

// Backward compatible methods (using old API)

// Tasks retrieves tasks with custom filters (backward compatible, deprecated)
// Recommended: use TasksList with NewQueryBuilder() instead
func (c *Client) Tasks(ctx context.Context, kwargs map[string]any) ([]models.Task, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// CRUD Operations

// TaskCreateParams contains parameters for creating a new task
type TaskCreateParams struct {
	Name        string   `json:"name"`
	ProjectID   string   `json:"project_id"`
	Text        string   `json:"text,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	Responsible string   `json:"responsible,omitempty"`
	Executors   []string `json:"executors,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Lists       []string `json:"lists,omitempty"` // sprints
	EpicID      string   `json:"epic_id,omitempty"`
	LogicTypeID string   `json:"logic_type_id,omitempty"`
}

// TaskCreate creates a new task
// Example:
//
//	params := evateamclient.TaskCreateParams{
//	  Name:      "New Task",
//	  ProjectID: "Project:uuid",
//	  Priority:  3,
//	}
//	task, err := client.TaskCreate(ctx, params)
func (c *Client) TaskCreate(
	ctx context.Context,
	params TaskCreateParams,
) (*models.Task, error) {
	kwargs := map[string]any{
		"name":       params.Name,
		"project_id": params.ProjectID,
	}

	if params.Text != "" {
		kwargs["text"] = params.Text
	}
	if params.Priority > 0 {
		kwargs["priority"] = params.Priority
	}
	if params.Deadline != "" {
		kwargs["deadline"] = params.Deadline
	}
	if params.Responsible != "" {
		kwargs["responsible"] = params.Responsible
	}
	if len(params.Executors) > 0 {
		kwargs["executors"] = params.Executors
	}
	if len(params.Tags) > 0 {
		kwargs["tags"] = params.Tags
	}
	if len(params.Lists) > 0 {
		kwargs["lists"] = params.Lists
	}
	if params.EpicID != "" {
		kwargs["epic_id"] = params.EpicID
	}
	if params.LogicTypeID != "" {
		kwargs["logic_type_id"] = params.LogicTypeID
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// TaskUpdate updates an existing task
// Example:
//
//	updates := map[string]any{
//	  "name":     "Updated Task Name",
//	  "priority": 5,
//	}
//	task, err := client.TaskUpdate(ctx, "CmfTask:uuid", updates)
func (c *Client) TaskUpdate(
	ctx context.Context,
	taskID string,
	updates map[string]any,
) (*models.Task, error) {
	kwargs := map[string]any{
		"id": taskID,
	}
	for k, v := range updates {
		kwargs[k] = v
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.update",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// TaskUpdateStatus updates task status (workflow transition)
// Example:
//
//	task, err := client.TaskUpdateStatus(ctx, "CmfTask:uuid", "CLOSED")
func (c *Client) TaskUpdateStatus(
	ctx context.Context,
	taskID string,
	status string,
) (*models.Task, error) {
	return c.TaskUpdate(ctx, taskID, map[string]any{
		"cache_status_type": status,
	})
}

// TaskDelete deletes a task by ID
// Example:
//
//	err := client.TaskDelete(ctx, "CmfTask:uuid")
func (c *Client) TaskDelete(
	ctx context.Context,
	taskID string,
) error {
	kwargs := map[string]any{
		"id": taskID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// TaskArchive archives a task (soft delete)
// Example:
//
//	err := client.TaskArchive(ctx, "CmfTask:uuid")
func (c *Client) TaskArchive(
	ctx context.Context,
	taskID string,
) error {
	_, err := c.TaskUpdate(ctx, taskID, map[string]any{
		"cmf_deleted": true,
	})
	return err
}
