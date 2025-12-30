package evateamclient

import (
	"context"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultSprintFields for sprint queries.
var DefaultSprintFields = []string{
	"id", "class_name", "code", "name", "parent", "cache_status_type",
	"start_date", "end_date", "goal",
}

// ProjectSprints retrieves ALL sprints for project.
func (c *Client) ProjectSprints(
	ctx context.Context,
	projectCode string,
	fields []string,
) ([]models.Sprint, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultSprintFields
	}
	kwargs := map[string]any{
		"filter": []any{"parent", "==", projectCode},
		"fields": fields,
	}

	return c.Sprints(ctx, kwargs)
}

func (c *Client) Sprint(
	ctx context.Context,
	sprintCode string,
	fields []string,
) (*models.Sprint, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultSprintFields
	}

	kwargs := map[string]any{
		"filter": []any{"code", "==", sprintCode},
		"fields": fields,
		"slice":  []int{0, 1},
	}

	sprints, meta, err := c.Sprints(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}

	if len(sprints) == 0 {
		return nil, meta, nil
	}

	return &sprints[0], meta, nil
}

// Sprints retrieves sprints with custom filters.
func (c *Client) Sprints(ctx context.Context, kwargs map[string]any) ([]models.Sprint, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}
	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}
	var resp models.SprintListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ActiveProjectSprint retrieves currently active sprint.
func (c *Client) ActiveProjectSprint(
	ctx context.Context,
	projectCode string,
) (*models.Sprint, *models.Meta, error) {
	kwargs := map[string]any{
		"filter": [][]any{
			{"parent", "==", projectCode},
			{"cache_status_type", "==", "OPEN"},
		},
		"fields": DefaultSprintFields,
		"slice":  []int{0, 1},
	}
	sprints, meta, err := c.Sprints(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}
	if len(sprints) == 0 {
		return nil, meta, nil
	}

	return &sprints[0], meta, nil
}
