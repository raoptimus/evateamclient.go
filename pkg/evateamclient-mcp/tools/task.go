package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

// TaskTools provides MCP tool handlers for task operations.
type TaskTools struct {
	client *evateamclient.Client
}

// NewTaskTools creates a new TaskTools instance.
func NewTaskTools(client *evateamclient.Client) *TaskTools {
	return &TaskTools{client: client}
}

// TaskListInput represents input for eva_task_list tool.
type TaskListInput struct {
	QueryInput

	// Optional project filter
	ProjectID string `json:"project_id,omitempty"`

	// Optional status filter
	StatusType string `json:"status_type,omitempty"`

	// Optional sprint/list filter
	SprintCode string `json:"sprint_code,omitempty"`

	// Optional responsible person filter
	ResponsibleID string `json:"responsible_id,omitempty"`

	// Optional logic type ID filter (e.g., task type like "Target", "Epic", etc.)
	LogicTypeID string `json:"logic_type_id,omitempty"`
}

// TaskList returns a list of tasks matching filters.
func (t *TaskTools) TaskList(ctx context.Context, input *TaskListInput) (*ListResult, error) {
	// Build kwargs for complex filters
	kwargs := BuildKwargs(&input.QueryInput)

	// Add specific filters
	var filters [][]any
	if existingFilter, ok := kwargs["filter"].([][]any); ok {
		filters = existingFilter
	} else {
		if singleFilter, ok := kwargs["filter"].([]any); ok {
			filters = [][]any{singleFilter}
		}
	}

	if input.ProjectID != "" {
		filters = append(filters, []any{"project_id", "==", input.ProjectID})
	}
	if input.StatusType != "" {
		filters = append(filters, []any{"cache_status_type", "==", input.StatusType})
	}
	if input.ResponsibleID != "" {
		filters = append(filters, []any{"responsible_id", "==", input.ResponsibleID})
	}
	if input.LogicTypeID != "" {
		filters = append(filters, []any{"logic_type_id", "==", input.LogicTypeID})
	}

	// Sprint uses "contains" operator
	if input.SprintCode != "" {
		filters = append(filters, []any{"lists", "contains", input.SprintCode})
	}

	if len(filters) == 1 {
		kwargs["filter"] = filters[0]
	} else if len(filters) > 1 {
		kwargs["filter"] = filters
	}

	// Set default fields if not specified
	if _, ok := kwargs["fields"]; !ok {
		kwargs["fields"] = evateamclient.DefaultTaskListFields
	}

	tasks, _, err := t.client.Tasks(ctx, kwargs)
	if err != nil {
		return nil, WrapError("task_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(tasks),
		HasMore: len(tasks) == input.Limit && input.Limit > 0,
	}, nil
}

// TaskGetInput represents input for eva_task_get tool.
type TaskGetInput struct {
	// Task code (e.g., "PROJ-123")
	Code string `json:"code,omitempty"`

	// Task ID (e.g., "CmfTask:uuid")
	ID string `json:"id,omitempty"`

	// Fields to return
	Fields []string `json:"fields,omitempty"`
}

// TaskGet retrieves a single task by code or ID.
func (t *TaskTools) TaskGet(ctx context.Context, input *TaskGetInput) (any, error) {
	var qb *evateamclient.QueryBuilder

	switch {
	case input.Code != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityTask).
			Where(sq.Eq{"code": input.Code}).
			Limit(1)
	case input.ID != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityTask).
			Where(sq.Eq{"id": input.ID}).
			Limit(1)
	default:
		return nil, WrapError("task_get", ErrInvalidInput)
	}

	task, _, err := t.client.TaskQuery(ctx, qb)
	if err != nil {
		return nil, WrapError("task_get", err)
	}

	return task, nil
}

// TaskCreateInput represents input for eva_task_create tool.
type TaskCreateInput struct {
	Name        string   `json:"name"`
	ProjectID   string   `json:"project_id"`
	Text        string   `json:"text,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	Responsible string   `json:"responsible,omitempty"`
	Executors   []string `json:"executors,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Lists       []string `json:"lists,omitempty"`
	EpicID      string   `json:"epic_id,omitempty"`
	LogicTypeID string   `json:"logic_type_id,omitempty"`
}

// TaskCreate creates a new task.
func (t *TaskTools) TaskCreate(ctx context.Context, input *TaskCreateInput) (any, error) {
	params := &evateamclient.TaskCreateParams{
		Name:        input.Name,
		ProjectID:   input.ProjectID,
		Text:        input.Text,
		Priority:    input.Priority,
		Deadline:    input.Deadline,
		Responsible: input.Responsible,
		Executors:   input.Executors,
		Tags:        input.Tags,
		Lists:       input.Lists,
		EpicID:      input.EpicID,
		LogicTypeID: input.LogicTypeID,
	}

	task, err := t.client.TaskCreate(ctx, params)
	if err != nil {
		return nil, WrapError("task_create", err)
	}

	return task, nil
}

// TaskUpdateInput represents input for eva_task_update tool.
type TaskUpdateInput struct {
	// Task ID (required)
	ID string `json:"id"`

	// Fields to update (any task field)
	Updates map[string]any `json:"updates"`
}

// TaskUpdate updates an existing task.
func (t *TaskTools) TaskUpdate(ctx context.Context, input TaskUpdateInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("task_update", ErrInvalidInput)
	}

	task, err := t.client.TaskUpdate(ctx, input.ID, input.Updates)
	if err != nil {
		return nil, WrapError("task_update", err)
	}

	return task, nil
}

// TaskUpdateStatusInput represents input for eva_task_update_status tool.
type TaskUpdateStatusInput struct {
	ID     string `json:"id"`
	Status string `json:"status"` // OPEN, IN_PROGRESS, CLOSED
}

// TaskUpdateStatus updates task status.
func (t *TaskTools) TaskUpdateStatus(ctx context.Context, input TaskUpdateStatusInput) (any, error) {
	if input.ID == "" || input.Status == "" {
		return nil, WrapError("task_update_status", ErrInvalidInput)
	}

	task, err := t.client.TaskUpdateStatus(ctx, input.ID, input.Status)
	if err != nil {
		return nil, WrapError("task_update_status", err)
	}

	return task, nil
}

// TaskDeleteInput represents input for eva_task_delete tool.
type TaskDeleteInput struct {
	ID string `json:"id"`
}

// TaskDelete deletes a task.
func (t *TaskTools) TaskDelete(ctx context.Context, input TaskDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("task_delete", ErrInvalidInput)
	}

	err := t.client.TaskDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("task_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// TaskArchiveInput represents input for eva_task_archive tool.
type TaskArchiveInput struct {
	ID string `json:"id"`
}

// TaskArchive archives a task (soft delete).
func (t *TaskTools) TaskArchive(ctx context.Context, input TaskArchiveInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("task_archive", ErrInvalidInput)
	}

	err := t.client.TaskArchive(ctx, input.ID)
	if err != nil {
		return nil, WrapError("task_archive", err)
	}

	return map[string]bool{"success": true}, nil
}

// TaskCountInput represents input for eva_task_count tool.
type TaskCountInput struct {
	ProjectID     string `json:"project_id,omitempty"`
	StatusType    string `json:"status_type,omitempty"`
	SprintCode    string `json:"sprint_code,omitempty"`
	ResponsibleID string `json:"responsible_id,omitempty"`
}

// TaskCount counts tasks matching filters.
func (t *TaskTools) TaskCount(ctx context.Context, input TaskCountInput) (*CountResult, error) {
	kwargs := make(map[string]any)
	var filters [][]any

	if input.ProjectID != "" {
		filters = append(filters, []any{"project_id", "==", input.ProjectID})
	}
	if input.StatusType != "" {
		filters = append(filters, []any{"cache_status_type", "==", input.StatusType})
	}
	if input.ResponsibleID != "" {
		filters = append(filters, []any{"responsible_id", "==", input.ResponsibleID})
	}
	if input.SprintCode != "" {
		filters = append(filters, []any{"lists", "contains", input.SprintCode})
	}

	if len(filters) == 1 {
		kwargs["filter"] = filters[0]
	} else if len(filters) > 1 {
		kwargs["filter"] = filters
	}

	count, _, err := t.client.TasksCount(ctx, kwargs)
	if err != nil {
		return nil, WrapError("task_count", err)
	}

	return &CountResult{Count: int(count)}, nil
}
