package evateamclient

import (
	"context"
	"fmt"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultEpicFields for epic queries.
var DefaultEpicFields = []string{"id", "code", "name", "project_id", "cache_status_type"}

// ProjectEpics retrieves ALL epics in project.
func (c *Client) ProjectEpics(
	ctx context.Context,
	projectCode string,
	fields []string,
) ([]models.Epic, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultEpicFields
	}

	kwargs := map[string]any{
		"filter": []any{"project_id", "==", fmt.Sprintf("Project:%s", projectCode)},
		"fields": fields,
	}

	return c.Epics(ctx, kwargs)
}

// EpicTasks retrieves ALL tasks in epic.
func (c *Client) EpicTasks(
	ctx context.Context,
	epicCode string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}

	kwargs := map[string]any{
		"filter": []any{"epic", "==", epicCode},
		"fields": fields,
	}

	return c.Tasks(ctx, kwargs)
}

// Epics retrieves epics with custom filters.
func (c *Client) Epics(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Epic, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Epic.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.EpicListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
