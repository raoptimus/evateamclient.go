package evateamclient

import (
	"context"
	"fmt"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultTaskFields for task queries.
var DefaultTaskFields = []string{
	"id", "code", "name", "project_id", "lists", "cmf_owner_id",
	"responsible", "cache_status_type", "priority", "deadline",
	"epic", "tags", "executors", "waiting_for", "parent_id",
	"fix_versions", "agile_story_points", "components", "logic_type",
}

// Task retrieves single task by code.
func (c *Client) Task(
	ctx context.Context,
	taskCode string,
	fields []string,
) (*models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}

	kwargs := map[string]any{
		"filter": []any{"code", "==", taskCode},
		"fields": fields,
		"slice":  []int{0, 1},
	}

	tasks, meta, err := c.Tasks(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}

	if len(tasks) == 0 {
		return nil, meta, nil
	}

	return &tasks[0], meta, nil
}

// ProjectTasks retrieves ALL tasks for project.
func (c *Client) ProjectTasks(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}
	kwargs := map[string]any{
		"filter": []any{"parent_id", "==", projectID},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// SprintTasks retrieves ALL tasks for sprint.
func (c *Client) SprintTasks(
	ctx context.Context,
	sprintCode string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}
	kwargs := map[string]any{
		"filter": []any{"lists", "contains", sprintCode},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// Tasks retrieves tasks with custom filters.
func (c *Client) Tasks(ctx context.Context, kwargs map[string]any) ([]models.Task, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
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

// PersonTasks retrieves ALL tasks assigned to user.
func (c *Client) PersonTasks(
	ctx context.Context,
	userID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}

	kwargs := map[string]any{
		"filter": [][]any{
			{"responsible", "==", userID},
			{"executors", "contains", userID},
		},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// PersonProjectTasks retrieves user's tasks in specific project.
func (c *Client) PersonProjectTasks(
	ctx context.Context,
	projectCode,
	userID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}

	kwargs := map[string]any{
		"filter": [][]any{
			{"project_id", "==", fmt.Sprintf("Project:%s", projectCode)},
			{"responsible", "==", userID},
			{"executors", "contains", userID},
		},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}
