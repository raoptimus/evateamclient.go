package evateamclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultTimeLogFields for time log queries.
var DefaultTimeLogFields = []string{
	"id", "task_id", "user_id", "user_name", "user_login",
	"minutes_spent", "date", "description", "cmf_created_at",
}

// TimeLog retrieves single time log entry by ID.
func (c *Client) TimeLog(ctx context.Context, timeLogID string, fields []string) (*models.CmfTaskTimeLog, *models.CmfMeta, error) {
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
func (c *Client) TaskTimeLogs(ctx context.Context, taskCode string, fields []string) ([]models.CmfTaskTimeLog, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter":   []any{"task_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// UserTaskTimeLogs retrieves time entries for task by specific user.
func (c *Client) UserTaskTimeLogs(ctx context.Context, taskCode, userID string, fields []string) ([]models.CmfTaskTimeLog, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter": [][]any{
			{"task_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
			{"user_id", "==", userID},
		},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// ProjectTimeLogs retrieves ALL time entries for project tasks.
func (c *Client) ProjectTimeLogs(ctx context.Context, projectCode string, fields []string) ([]models.CmfTaskTimeLog, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTimeLogFields
	}

	kwargs := map[string]any{
		"filter":   []any{"task_id.project_id", "==", fmt.Sprintf("CmfProject:%s", projectCode)},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TimeLogs(ctx, kwargs)
}

// TimeLogs retrieves time logs with custom filters.
func (c *Client) TimeLogs(ctx context.Context, kwargs map[string]any) ([]models.CmfTaskTimeLog, *models.CmfMeta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTaskTimeLog.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfTaskTimeLogListResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfTaskTimeLog.list", reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
