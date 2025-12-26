package evateamclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raoptimus/evateamclient/models"
)

// TasksCount returns total tasks matching filters.
func (c *Client) TasksCount(ctx context.Context, kwargs map[string]any) (int64, *models.CmfMeta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := rpcRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfCountResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfTask.count", reqBody, &resp); err != nil {
		return 0, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ProjectTasksCount returns total tasks in project.
func (c *Client) ProjectTasksCount(ctx context.Context, projectCode string) (int64, *models.CmfMeta, error) {
	kwargs := map[string]any{
		"filter": []any{"project_id", "==", fmt.Sprintf("CmfProject:%s", projectCode)},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintTasksCount returns total tasks in sprint.
func (c *Client) SprintTasksCount(ctx context.Context, sprintCode string) (int64, *models.CmfMeta, error) {
	kwargs := map[string]any{
		"filter": []any{"lists", "contains", sprintCode},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintStats retrieves sprint statistics.
func (c *Client) SprintStats(ctx context.Context, sprintCode string) (*models.SprintStats, error) {
	// Implementation via aggregation queries or custom API method
	// For now, calculate from tasks + time logs
	tasks, _, err := c.SprintTasks(ctx, sprintCode, []string{"cache_status_type"})
	if err != nil {
		return nil, err
	}

	stats := &models.SprintStats{
		SprintID:   sprintCode,
		TotalTasks: len(tasks),
	}

	// Count by status
	statusCount := make(map[string]int)
	for _, task := range tasks {
		statusCount[task.CacheStatus]++
	}
	stats.TasksByStatus = statusCount

	return stats, nil
}

// ProjectStats retrieves project statistics.
func (c *Client) ProjectStats(ctx context.Context, projectCode string) (*models.ProjectStats, *models.CmfMeta, error) {
	stats := &models.ProjectStats{ProjectID: projectCode}

	// Total tasks
	count, _, err := c.ProjectTasksCount(ctx, projectCode)
	if err == nil {
		stats.TotalTasks = int(count)
	}

	// Open tasks
	openCount, _, err := c.TasksCount(ctx, map[string]any{
		"filter": [][]any{
			{"project_id", "==", fmt.Sprintf("CmfProject:%s", projectCode)},
			{"cache_status_type", "==", "OPEN"},
		},
	})
	if err == nil {
		stats.OpenTasks = int(openCount)
	}

	// Active sprints
	sprints, _, err := c.ProjectSprints(ctx, projectCode, []string{"cache_status_type"})
	if err == nil {
		for _, sprint := range sprints {
			if sprint.CacheStatus == "OPEN" {
				stats.ActiveSprints++
			}
		}
	}

	// Total users
	users, _, err := c.ProjectUsers(ctx, projectCode, nil)
	if err == nil {
		stats.TotalUsers = len(users)
	}

	return stats, nil, nil
}
