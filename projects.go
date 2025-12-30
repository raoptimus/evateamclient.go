package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Project field constants for type-safe queries
const (
	ProjectFieldID                     = "id"
	ProjectFieldClassName              = "class_name"
	ProjectFieldCode                   = "code"
	ProjectFieldName                   = "name"
	ProjectFieldText                   = "text"
	ProjectFieldCMFLockedAt            = "cmf_locked_at"
	ProjectFieldCMFCreatedAt           = "cmf_created_at"
	ProjectFieldCMFModifiedAt          = "cmf_modified_at"
	ProjectFieldCMFViewedAt            = "cmf_viewed_at"
	ProjectFieldCMFDeleted             = "cmf_deleted"
	ProjectFieldCMFVersion             = "cmf_version"
	ProjectFieldCacheStatusType        = "cache_status_type"
	ProjectFieldWorkflowType           = "workflow_type"
	ProjectFieldWorkflowID             = "workflow_id"
	ProjectFieldParentID               = "parent_id"
	ProjectFieldCmfOwnerID             = "cmf_owner_id"
	ProjectFieldSystem                 = "system"
	ProjectFieldImportOriginal         = "import_original"
	ProjectFieldSlOwnerLock            = "sl_owner_lock"
	ProjectFieldPermParentOwnerID      = "perm_parent_owner_id"
	ProjectFieldPermInheritACLID       = "perm_inherit_acl_id"
	ProjectFieldPermEffectiveACLID     = "perm_effective_acl_id"
	ProjectFieldPermSecurityLevelCache = "perm_security_level_allowed_ids_cache"
	ProjectFieldIsTemplate             = "is_template"
	ProjectFieldExecutors              = "executors"
	ProjectFieldAdmins                 = "cmfprojectadmins"
	ProjectFieldSpectators             = "spectators"
	ProjectFieldOwnerAssistants        = "cmf_owner_assistants"
)

var (
	// DefaultProjectFields - standard projection for single project queries
	DefaultProjectFields = []string{
		ProjectFieldID,
		ProjectFieldName,
		ProjectFieldCode,
		ProjectFieldCMFCreatedAt,
		ProjectFieldCMFModifiedAt,
		ProjectFieldExecutors,
		ProjectFieldAdmins,
		ProjectFieldSpectators,
		ProjectFieldOwnerAssistants,
	}

	// DefaultProjectListFields - optimized for LIST queries (lighter payload)
	DefaultProjectListFields = []string{
		ProjectFieldID,
		ProjectFieldClassName,
		ProjectFieldCode,
		ProjectFieldName,
		ProjectFieldCacheStatusType,
		ProjectFieldCmfOwnerID,
		ProjectFieldWorkflowID,
		ProjectFieldSystem,
		ProjectFieldSlOwnerLock,
	}
)

// Project retrieves a single project by code (backward compatible)
// Example:
//
//	project, meta, err := client.Project(ctx, "PROJ-123", nil)
func (c *Client) Project(
	ctx context.Context,
	code string,
	fields []string,
) (*models.Project, *models.Meta, error) {
	// Use real Squirrel builder
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityProject).
		Where(sq.Eq{ProjectFieldCode: code}).
		Limit(1)

	return c.ProjectQuery(ctx, qb)
}

// ProjectQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "name", "executors").
//	  From(evateamclient.EntityProject).
//	  Where(sq.Eq{"code": "PROJ-123"})
//	project, meta, err := client.ProjectQuery(ctx, qb)
func (c *Client) ProjectQuery(
	ctx context.Context,
	qb *QueryBuilder,
) (*models.Project, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultProjectFields
	}

	method, err := qb.ToMethod()
	if err != nil {
		return nil, nil, err
	}

	// Force .get method for single result
	method = "CmfProject.get"

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ProjectGetResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// ProjectsList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityProject).
//	  Where(sq.Like{"name": "%Mobile%"}).
//	  Where(sq.Eq{"system": false}).
//	  OrderBy("-cmf_created_at").
//	  Offset(0).Limit(50)
//	projects, meta, err := client.ProjectsList(ctx, qb)
func (c *Client) ProjectsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Project, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultProjectListFields
	}

	method, err := qb.ToMethod()
	if err != nil {
		return nil, nil, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ProjectListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ProjectCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityProject).
//	  Where(sq.Eq{"system": false})
//	count, err := client.ProjectCount(ctx, qb)
func (c *Client) ProjectCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfProject.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  int    `json:"result"`
	}

	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// Backward compatible methods (using old API)

// Projects retrieves a list of projects (backward compatible, deprecated)
// Recommended: use ProjectsSquirrel with NewEvaBuilder() instead
func (c *Client) Projects(
	ctx context.Context,
	fields []string,
	kwargs map[string]any,
) ([]models.Project, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if len(fields) == 0 {
		fields = DefaultProjectListFields
	}
	kwargs["fields"] = fields

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfProject.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ProjectListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
