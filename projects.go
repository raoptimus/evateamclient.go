package evateamclient

import (
	"context"
	"net/http"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultProjectFields are the recommended base fields for most use cases.
var DefaultProjectFields = []string{
	"id", "class_name", "code", "name", "cache_status_type",
	"workflow_type", "parent_id", "project_id", "cmf_owner_id", "workflow_id",
	"system", "sl_owner_lock", "is_template",
}

// DefaultProjectListFields are recommended fields for project list queries.
var DefaultProjectListFields = []string{
	"id", "class_name", "code", "name", "cache_status_type",
	"cmf_owner_id", "workflow_id", "system", "sl_owner_lock",
}

// Project retrieves a project by code with optional field selection.
// If fields is nil or empty, uses DefaultProjectFields.
func (c *Client) Project(ctx context.Context, code string, fields []string) (*models.CmfProject, *models.CmfMeta, error) {
	kwargs := map[string]any{
		"filter": []any{"code", "==", code},
	}

	// Use default fields if none specified
	if len(fields) == 0 {
		fields = DefaultProjectFields
	}
	kwargs["fields"] = fields

	reqBody := rpcRequest{
		JSONRPC: "2.2",
		Method:  "CmfProject.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfProjectGetResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfProject.get", reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// ProjectFull retrieves project with ALL available fields (no fields filter).
func (c *Client) ProjectFull(ctx context.Context, code string) (*models.CmfProject, *models.CmfMeta, error) {
	return c.Project(ctx, code, nil) // nil fields = server default (usually all)
}

// Projects retrieves list of projects with optional field selection and filters.
func (c *Client) Projects(ctx context.Context, fields []string, kwargs map[string]any) ([]models.CmfProject, *models.CmfMeta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if len(fields) == 0 {
		fields = DefaultProjectListFields
	}
	kwargs["fields"] = fields

	reqBody := rpcRequest{
		JSONRPC: "2.2",
		Method:  "CmfProject.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfProjectListResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfProject.list", reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
