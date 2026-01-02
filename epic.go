package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Epic is a CmfTask with logic_type.code = "task.epic"
// Use TaskField* constants for field names

const (
	// LogicTypeEpic is the logic_type.code for epics
	LogicTypeEpic = "task.epic"
)

var (
	// DefaultEpicFields - standard projection for epic queries
	DefaultEpicFields = []string{
		TaskFieldID,
		TaskFieldClassName,
		TaskFieldCode,
		TaskFieldName,
		TaskFieldText,
		TaskFieldProjectID,
		TaskFieldCacheStatusType,
		TaskFieldLogicType,
	}

	// DefaultEpicListFields - optimized for LIST queries
	DefaultEpicListFields = []string{
		TaskFieldID,
		TaskFieldCode,
		TaskFieldName,
		TaskFieldProjectID,
		TaskFieldCacheStatusType,
	}
)

// Epic retrieves a single epic by code
// Note: Epics are CmfTask with logic_type.code = "task.epic"
// Example:
//
//	epic, meta, err := client.Epic(ctx, "UDMP-123", nil)
func (c *Client) Epic(
	ctx context.Context,
	epicCode string,
	fields []string,
) (*models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultEpicFields
	}

	kwargs := map[string]any{
		"filter": []any{TaskFieldCode, "==", epicCode},
		"fields": fields,
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

// EpicByID retrieves a single epic by ID
// Example:
//
//	epic, meta, err := client.EpicByID(ctx, "CmfTask:uuid", nil)
func (c *Client) EpicByID(
	ctx context.Context,
	epicID string,
	fields []string,
) (*models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultEpicFields
	}

	kwargs := map[string]any{
		"id":     epicID,
		"fields": fields,
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

// ProjectEpics retrieves ALL epics in project
// Example:
//
//	epics, meta, err := client.ProjectEpics(ctx, "CmfProject:uuid", nil)
func (c *Client) ProjectEpics(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultEpicListFields
	}

	kwargs := map[string]any{
		"filter": [][]any{
			{TaskFieldProjectID, "==", projectID},
			{"logic_type.code", "==", LogicTypeEpic},
		},
		"fields": fields,
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

// EpicTasks retrieves ALL tasks in epic by epic ID
// Example:
//
//	tasks, meta, err := client.EpicTasks(ctx, "CmfTask:uuid", nil)
func (c *Client) EpicTasks(
	ctx context.Context,
	epicID string,
	fields []string,
) ([]models.Task, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldEpicID: epicID})

	return c.TasksList(ctx, qb)
}

// Epics retrieves epics with custom kwargs
// Example:
//
//	epics, meta, err := client.Epics(ctx, map[string]any{
//	  "filter": [][]any{
//	    {"project_id", "==", "CmfProject:uuid"},
//	    {"logic_type.code", "==", "task.epic"},
//	  },
//	})
func (c *Client) Epics(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Task, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultEpicListFields
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
