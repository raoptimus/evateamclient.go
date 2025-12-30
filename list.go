package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// List field constants for type-safe queries
const (
	// Core fields
	ListFieldID                = "id"
	ListFieldClassName         = "class_name"
	ListFieldCode              = "code"
	ListFieldName              = "name"
	ListFieldCacheStatusType   = "cache_status_type"
	ListFieldCacheMembersCount = "cache_members_count"
	ListFieldLimitDays         = "limit_days"

	// Relations
	ListFieldParent     = "parent" // nested project object
	ListFieldParentID   = "parent_id"
	ListFieldProjectID  = "project_id"
	ListFieldWorkflowID = "workflow_id"

	// Date fields
	ListFieldStartDate = "start_date"
	ListFieldEndDate   = "end_date"

	// Content
	ListFieldGoal = "goal"
	ListFieldText = "text"

	// System
	ListFieldSystem        = "system"
	ListFieldSlOwnerLock   = "sl_owner_lock"
	ListFieldCmfOwnerID    = "cmf_owner_id"
	ListFieldCmfCreatedAt  = "cmf_created_at"
	ListFieldCmfModifiedAt = "cmf_modified_at"

	// Code prefixes for list types
	ListCodePrefixSprint  = "SPR-"
	ListCodePrefixRelease = "REL-"
)

var (
	// DefaultListFields - standard projection for single list queries
	DefaultListFields = []string{
		ListFieldID,
		ListFieldClassName,
		ListFieldCode,
		ListFieldName,
		ListFieldCacheStatusType,
		ListFieldCacheMembersCount,
		ListFieldLimitDays,
		ListFieldParent,
		ListFieldParentID,
		ListFieldProjectID,
		ListFieldCmfOwnerID,
		ListFieldWorkflowID,
		ListFieldStartDate,
		ListFieldEndDate,
		ListFieldGoal,
	}

	// DefaultListListFields - optimized for LIST queries (lighter payload)
	DefaultListListFields = []string{
		ListFieldID,
		ListFieldClassName,
		ListFieldCode,
		ListFieldName,
		ListFieldCacheStatusType,
		ListFieldCacheMembersCount,
		ListFieldParent,
		ListFieldParentID,
		ListFieldProjectID,
		ListFieldCmfOwnerID,
		ListFieldWorkflowID,
	}
)

// List retrieves a single list (sprint/release) by code
// Example:
//
//	list, meta, err := client.List(ctx, "SPR-001543", nil)
//	list, meta, err := client.List(ctx, "REL-001641", nil)
func (c *Client) List(
	ctx context.Context,
	listCode string,
	fields []string,
) (*models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldCode: listCode}).
		Limit(1)

	return c.ListQuery(ctx, qb)
}

// ListQuery executes query using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "name", "code", "cache_status_type").
//	  From(evateamclient.EntityList).
//	  Where(sq.Eq{"code": "SPR-001543"})
//	list, meta, err := client.ListQuery(ctx, qb)
func (c *Client) ListQuery(ctx context.Context, qb *QueryBuilder) (*models.List, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// ListsList retrieves lists using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name", "cache_status_type").
//	  From(evateamclient.EntityList).
//	  Where(sq.Eq{"project_id": "CmfProject:uuid"}).
//	  Where(sq.Eq{"cache_status_type": "OPEN"}).
//	  Offset(0).Limit(50)
//	lists, meta, err := client.ListsList(ctx, qb)
func (c *Client) ListsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.List, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultListListFields
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

	var resp models.ListListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ListCount counts lists using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityList).
//	  Where(sq.Eq{"project_id": "CmfProject:uuid"})
//	count, err := client.ListCount(ctx, qb)
func (c *Client) ListCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.count",
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

// ProjectLists retrieves ALL lists (sprints + releases) for project
// Example:
//
//	lists, meta, err := client.ProjectLists(ctx, "CmfProject:uuid", nil)
func (c *Client) ProjectLists(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID})

	return c.ListsList(ctx, qb)
}

// OpenProjectLists retrieves all open lists for project
// Example:
//
//	lists, meta, err := client.OpenProjectLists(ctx, "CmfProject:uuid", nil)
func (c *Client) OpenProjectLists(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Where(sq.Eq{ListFieldCacheStatusType: "OPEN"})

	return c.ListsList(ctx, qb)
}

// Lists retrieves lists with custom kwargs
// Example:
//
//	lists, meta, err := client.Lists(ctx, map[string]any{
//	  "filter": []any{"project_id", "==", "CmfProject:uuid"},
//	})
func (c *Client) Lists(ctx context.Context, kwargs map[string]any) ([]models.List, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultListListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ListListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// CRUD Operations

// ListCreateParams contains parameters for creating a new list (sprint/release)
type ListCreateParams struct {
	Name      string `json:"name"`
	ParentID  string `json:"parent_id"` // project ID (CmfProject:uuid)
	Code      string `json:"code,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

// ListCreate creates a new list (sprint/release)
// Example:
//
//	params := evateamclient.ListCreateParams{
//	  Name:     "Sprint 1",
//	  ParentID: "CmfProject:uuid",
//	}
//	list, err := client.ListCreate(ctx, params)
func (c *Client) ListCreate(
	ctx context.Context,
	params ListCreateParams,
) (*models.List, error) {
	kwargs := map[string]any{
		"name":      params.Name,
		"parent_id": params.ParentID,
	}

	if params.Code != "" {
		kwargs["code"] = params.Code
	}
	if params.StartDate != "" {
		kwargs["start_date"] = params.StartDate
	}
	if params.EndDate != "" {
		kwargs["end_date"] = params.EndDate
	}
	if params.Goal != "" {
		kwargs["goal"] = params.Goal
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// ListUpdate updates an existing list
// Example:
//
//	updates := map[string]any{
//	  "name": "Updated Name",
//	  "goal": "Complete feature X",
//	}
//	list, err := client.ListUpdate(ctx, "CmfList:uuid", updates)
func (c *Client) ListUpdate(
	ctx context.Context,
	listID string,
	updates map[string]any,
) (*models.List, error) {
	kwargs := map[string]any{
		"id": listID,
	}
	for k, v := range updates {
		kwargs[k] = v
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.update",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.ListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// ListClose closes a list (sprint/release)
// Example:
//
//	list, err := client.ListClose(ctx, "CmfList:uuid")
func (c *Client) ListClose(
	ctx context.Context,
	listID string,
) (*models.List, error) {
	return c.ListUpdate(ctx, listID, map[string]any{
		"cache_status_type": "CLOSED",
	})
}

// ListDelete deletes a list by ID
// Example:
//
//	err := client.ListDelete(ctx, "CmfList:uuid")
func (c *Client) ListDelete(
	ctx context.Context,
	listID string,
) error {
	kwargs := map[string]any{
		"id": listID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfList.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// =============================================================================
// Sprint-specific methods (code starts with "SPR-")
// =============================================================================

// ProjectSprints retrieves all sprints for project (code like "SPR-%")
// Example:
//
//	sprints, meta, err := client.ProjectSprints(ctx, "CmfProject:uuid", nil)
func (c *Client) ProjectSprints(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Where(sq.Like{ListFieldCode: ListCodePrefixSprint + "%"})

	return c.ListsList(ctx, qb)
}

// OpenProjectSprints retrieves open sprints for project
// Example:
//
//	sprints, meta, err := client.OpenProjectSprints(ctx, "CmfProject:uuid", nil)
func (c *Client) OpenProjectSprints(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Where(sq.Like{ListFieldCode: ListCodePrefixSprint + "%"}).
		Where(sq.Eq{ListFieldCacheStatusType: "OPEN"})

	return c.ListsList(ctx, qb)
}

// Sprints retrieves sprints with custom kwargs
// Example:
//
//	sprints, meta, err := client.Sprints(ctx, map[string]any{
//	  "filter": [][]any{
//	    {"project_id", "==", "CmfProject:uuid"},
//	    {"code", "LIKE", "SPR-%"},
//	  },
//	})
func (c *Client) Sprints(ctx context.Context, kwargs map[string]any) ([]models.List, *models.Meta, error) {
	return c.Lists(ctx, kwargs)
}

// =============================================================================
// Release-specific methods (code starts with "REL-")
// =============================================================================

// ProjectReleases retrieves all releases for project (code like "REL-%")
// Example:
//
//	releases, meta, err := client.ProjectReleases(ctx, "CmfProject:uuid", nil)
func (c *Client) ProjectReleases(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Where(sq.Like{ListFieldCode: ListCodePrefixRelease + "%"})

	return c.ListsList(ctx, qb)
}

// OpenProjectReleases retrieves open releases for project
// Example:
//
//	releases, meta, err := client.OpenProjectReleases(ctx, "CmfProject:uuid", nil)
func (c *Client) OpenProjectReleases(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.List, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Where(sq.Like{ListFieldCode: ListCodePrefixRelease + "%"}).
		Where(sq.Eq{ListFieldCacheStatusType: "OPEN"})

	return c.ListsList(ctx, qb)
}

// Releases retrieves releases with custom kwargs
// Example:
//
//	releases, meta, err := client.Releases(ctx, map[string]any{
//	  "filter": [][]any{
//	    {"project_id", "==", "CmfProject:uuid"},
//	    {"code", "LIKE", "REL-%"},
//	  },
//	})
func (c *Client) Releases(ctx context.Context, kwargs map[string]any) ([]models.List, *models.Meta, error) {
	return c.Lists(ctx, kwargs)
}
