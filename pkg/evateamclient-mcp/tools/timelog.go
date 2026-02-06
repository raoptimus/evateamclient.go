package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

// TimeLogTools provides MCP tool handlers for time log operations.
type TimeLogTools struct {
	client *evateamclient.Client
}

// NewTimeLogTools creates a new TimeLogTools instance.
func NewTimeLogTools(client *evateamclient.Client) *TimeLogTools {
	return &TimeLogTools{client: client}
}

// TimeLogListInput represents input for eva_timelog_list tool.
type TimeLogListInput struct {
	QueryInput
	TaskID    string `json:"task_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// TimeLogList returns a list of time log entries.
func (t *TimeLogTools) TimeLogList(ctx context.Context, input *TimeLogListInput) (*ListResult, error) {
	// For project filter, use kwargs (dot notation for nested field)
	if input.ProjectID != "" {
		kwargs := BuildKwargs(&input.QueryInput)
		kwargs["filter"] = []any{"parent.project_id", "==", input.ProjectID}

		logs, _, err := t.client.TimeLogs(ctx, kwargs)
		if err != nil {
			return nil, WrapError("timelog_list", err)
		}

		return &ListResult{
			Items:   toAnySlice(logs),
			HasMore: len(logs) == input.Limit && input.Limit > 0,
		}, nil
	}

	qb, err := BuildQuery(evateamclient.EntityTimeLog, &input.QueryInput)
	if err != nil {
		return nil, WrapError("timelog_list", err)
	}

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"parent_id": input.TaskID})
	}
	if input.UserID != "" {
		qb = qb.Where(sq.Eq{"cmf_owner_id": input.UserID})
	}

	logs, _, err := t.client.TimeLogsList(ctx, qb)
	if err != nil {
		return nil, WrapError("timelog_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(logs),
		HasMore: len(logs) == input.Limit && input.Limit > 0,
	}, nil
}

// TimeLogGetInput represents input for eva_timelog_get tool.
type TimeLogGetInput struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields,omitempty"`
}

// TimeLogGet retrieves a single time log entry.
func (t *TimeLogTools) TimeLogGet(ctx context.Context, input TimeLogGetInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("timelog_get", ErrInvalidInput)
	}

	log, _, err := t.client.TimeLog(ctx, input.ID, input.Fields)
	if err != nil {
		return nil, WrapError("timelog_get", err)
	}

	return log, nil
}

// TimeLogCreateInput represents input for eva_timelog_create tool.
type TimeLogCreateInput struct {
	TaskID    string `json:"task_id"`
	TimeSpent int    `json:"time_spent"` // minutes
}

// TimeLogCreate creates a new time log entry.
func (t *TimeLogTools) TimeLogCreate(ctx context.Context, input TimeLogCreateInput) (any, error) {
	if input.TaskID == "" || input.TimeSpent <= 0 {
		return nil, WrapError("timelog_create", ErrInvalidInput)
	}

	params := evateamclient.TimeLogCreateParams{
		ParentID:  input.TaskID,
		TimeSpent: input.TimeSpent,
	}

	log, err := t.client.TimeLogCreate(ctx, params)
	if err != nil {
		return nil, WrapError("timelog_create", err)
	}

	return log, nil
}

// TimeLogUpdateInput represents input for eva_timelog_update tool.
type TimeLogUpdateInput struct {
	ID      string         `json:"id"`
	Updates map[string]any `json:"updates"`
}

// TimeLogUpdate updates an existing time log entry.
func (t *TimeLogTools) TimeLogUpdate(ctx context.Context, input TimeLogUpdateInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("timelog_update", ErrInvalidInput)
	}

	log, err := t.client.TimeLogUpdate(ctx, input.ID, input.Updates)
	if err != nil {
		return nil, WrapError("timelog_update", err)
	}

	return log, nil
}

// TimeLogDeleteInput represents input for eva_timelog_delete tool.
type TimeLogDeleteInput struct {
	ID string `json:"id"`
}

// TimeLogDelete deletes a time log entry.
func (t *TimeLogTools) TimeLogDelete(ctx context.Context, input TimeLogDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("timelog_delete", ErrInvalidInput)
	}

	err := t.client.TimeLogDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("timelog_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// TimeLogCountInput represents input for eva_timelog_count tool.
type TimeLogCountInput struct {
	TaskID string `json:"task_id,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// TimeLogCount counts time log entries.
func (t *TimeLogTools) TimeLogCount(ctx context.Context, input TimeLogCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityTimeLog)

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"parent_id": input.TaskID})
	}
	if input.UserID != "" {
		qb = qb.Where(sq.Eq{"cmf_owner_id": input.UserID})
	}

	count, err := t.client.TimeLogCount(ctx, qb)
	if err != nil {
		return nil, WrapError("timelog_count", err)
	}

	return &CountResult{Count: count}, nil
}
