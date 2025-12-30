package evateamclient

import (
	"context"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultTimeLogFields for time log queries.
var DefaultTimeLogFields = []string{
	"id", "code", "time_spent", "author", "parent",
	"description", "cmf_created_at", "cmf_modified_at",
}

// TimeLog retrieves single time log entry by ID.
func (c *Client) TimeLog(
	ctx context.Context,
	timeLogID string,
	fields []string,
) (*models.TaskTimeLog, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter": []any{"id", "==", timeLogID},
		"fields": fields,
		"slice":  []int{0, 1},
	}

	logs, meta, err := c.TimeLogs(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}

	if len(logs) == 0 {
		return nil, meta, nil
	}

	return &logs[0], meta, nil
}

// TaskTimeLogs retrieves ALL time entries for specific task.
func (c *Client) TaskTimeLogs(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskTimeLog, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter":   []any{"parent", "==", taskID},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// UserTaskTimeLogs retrieves time entries for task by specific user.
func (c *Client) UserTaskTimeLogs(
	ctx context.Context,
	taskID,
	userID string,
	fields []string,
) ([]models.TaskTimeLog, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter": [][]any{
			{"task_id", "==", taskID},
			{"user_id", "==", userID},
		},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// ProjectTimeLogs retrieves ALL time entries for project tasks.
func (c *Client) ProjectTimeLogs(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.TaskTimeLog, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter":   []any{"task_id.project_id", "==", projectID},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// TimeLogs retrieves time logs with custom filters.
func (c *Client) TimeLogs(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.TaskTimeLog, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTimeTrackerHistory.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskTimeLogListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
