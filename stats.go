package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// TasksCount returns total tasks matching filters.
func (c *Client) TasksCount(ctx context.Context, kwargs map[string]any) (int64, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTask.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CountResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ProjectTasksCount returns total tasks in project.
func (c *Client) ProjectTasksCount(ctx context.Context, projectID string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldProjectID, "==", projectID},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintTasksCount returns total tasks in sprint by list code.
func (c *Client) SprintTasksCount(ctx context.Context, sprintCode string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldLists, "contains", sprintCode},
	}
	return c.TasksCount(ctx, kwargs)
}

// ListTasksCount returns total tasks in list (sprint/release) by list code.
func (c *Client) ListTasksCount(ctx context.Context, listCode string) (int64, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": []any{TaskFieldLists, "contains", listCode},
	}
	return c.TasksCount(ctx, kwargs)
}

// SprintStats retrieves sprint statistics.
func (c *Client) SprintStats(ctx context.Context, sprintCode string) (*models.SprintStats, error) {
	tasks, _, err := c.SprintTasks(ctx, sprintCode, []string{TaskFieldCacheStatusType})
	if err != nil {
		return nil, err
	}

	stats := &models.SprintStats{
		SprintID:   sprintCode,
		TotalTasks: len(tasks),
	}

	statusCount := make(map[string]int)
	for _, task := range tasks {
		statusCount[task.CacheStatusType]++
	}
	stats.TasksByStatus = statusCount

	return stats, nil
}

// ProjectStats retrieves project statistics.
func (c *Client) ProjectStats(ctx context.Context, projectID string) (*models.ProjectStats, *models.Meta, error) {
	stats := &models.ProjectStats{ProjectID: projectID}

	// Total tasks
	count, _, err := c.ProjectTasksCount(ctx, projectID)
	if err == nil {
		stats.TotalTasks = int(count)
	}

	// Open tasks
	qb := NewQueryBuilder().
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Where(sq.Eq{TaskFieldCacheStatusType: StatusTypeOpen})
	openCount, err := c.TaskCount(ctx, qb)
	if err == nil {
		stats.OpenTasks = openCount
	}

	// Active sprints
	sprints, _, err := c.OpenProjectSprints(ctx, projectID, []string{ListFieldID})
	if err == nil {
		stats.ActiveSprints = len(sprints)
	}

	// Total users (from project executors)
	qb = NewQueryBuilder().
		Select("executors").
		From(EntityProject).
		Where(sq.Eq{ProjectFieldID: projectID}).
		Limit(1)
	project, _, err := c.ProjectQuery(ctx, qb)
	if err == nil && project != nil {
		stats.TotalUsers = len(project.Executors)
	}

	return stats, nil, nil
}
