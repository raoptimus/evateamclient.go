package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

const (
	listTypeSprint  = "sprint"
	listTypeRelease = "release"
)

// ListTools provides MCP tool handlers for list (sprint/release) operations.
type ListTools struct {
	client *evateamclient.Client
}

// NewListTools creates a new ListTools instance.
func NewListTools(client *evateamclient.Client) *ListTools {
	return &ListTools{client: client}
}

// ListListInput represents input for eva_list_list tool.
type ListListInput struct {
	QueryInput

	// Filter by project
	ProjectID string `json:"project_id,omitempty"`

	// Filter by status
	StatusType string `json:"status_type,omitempty"`

	// Filter by type: "sprint" or "release"
	Type string `json:"type,omitempty"`
}

// ListList returns a list of lists (sprints/releases).
func (l *ListTools) ListList(ctx context.Context, input *ListListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityList, &input.QueryInput)
	if err != nil {
		return nil, WrapError("list_list", err)
	}

	// Add project filter
	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}

	// Add status filter
	if input.StatusType != "" {
		qb = qb.Where(sq.Eq{"cache_status_type": input.StatusType})
	}

	// Add type filter (sprint/release based on code prefix)
	switch input.Type {
	case listTypeSprint:
		qb = qb.Where(sq.Like{"code": evateamclient.ListCodePrefixSprint + "%"})
	case listTypeRelease:
		qb = qb.Where(sq.Like{"code": evateamclient.ListCodePrefixRelease + "%"})
	}

	lists, _, err := l.client.ListsList(ctx, qb)
	if err != nil {
		return nil, WrapError("list_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(lists),
		HasMore: len(lists) == input.Limit && input.Limit > 0,
	}, nil
}

// ListGetInput represents input for eva_list_get tool.
type ListGetInput struct {
	// List code (e.g., "SPR-001543", "REL-001641")
	Code string `json:"code,omitempty"`

	// List ID (e.g., "CmfList:uuid")
	ID string `json:"id,omitempty"`

	// Fields to return
	Fields []string `json:"fields,omitempty"`
}

// ListGet retrieves a single list by code or ID.
func (l *ListTools) ListGet(ctx context.Context, input *ListGetInput) (any, error) {
	var qb *evateamclient.QueryBuilder

	switch {
	case input.Code != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityList).
			Where(sq.Eq{"code": input.Code}).
			Limit(1)
	case input.ID != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityList).
			Where(sq.Eq{"id": input.ID}).
			Limit(1)
	default:
		return nil, WrapError("list_get", ErrInvalidInput)
	}

	list, _, err := l.client.ListQuery(ctx, qb)
	if err != nil {
		return nil, WrapError("list_get", err)
	}

	return list, nil
}

// ListCreateInput represents input for eva_list_create tool.
type ListCreateInput struct {
	Name      string `json:"name"`
	ParentID  string `json:"parent_id"` // Project ID
	Code      string `json:"code,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

// ListCreate creates a new list (sprint/release).
func (l *ListTools) ListCreate(ctx context.Context, input *ListCreateInput) (any, error) {
	params := &evateamclient.ListCreateParams{
		Name:      input.Name,
		ParentID:  input.ParentID,
		Code:      input.Code,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Goal:      input.Goal,
	}

	list, err := l.client.ListCreate(ctx, params)
	if err != nil {
		return nil, WrapError("list_create", err)
	}

	return list, nil
}

// ListUpdateInput represents input for eva_list_update tool.
type ListUpdateInput struct {
	ID      string         `json:"id"`
	Updates map[string]any `json:"updates"`
}

// ListUpdate updates an existing list.
func (l *ListTools) ListUpdate(ctx context.Context, input ListUpdateInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("list_update", ErrInvalidInput)
	}

	list, err := l.client.ListUpdate(ctx, input.ID, input.Updates)
	if err != nil {
		return nil, WrapError("list_update", err)
	}

	return list, nil
}

// ListCloseInput represents input for eva_list_close tool.
type ListCloseInput struct {
	ID string `json:"id"`
}

// ListClose closes a list (sprint/release).
func (l *ListTools) ListClose(ctx context.Context, input ListCloseInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("list_close", ErrInvalidInput)
	}

	list, err := l.client.ListClose(ctx, input.ID)
	if err != nil {
		return nil, WrapError("list_close", err)
	}

	return list, nil
}

// ListDeleteInput represents input for eva_list_delete tool.
type ListDeleteInput struct {
	ID string `json:"id"`
}

// ListDelete deletes a list.
func (l *ListTools) ListDelete(ctx context.Context, input ListDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("list_delete", ErrInvalidInput)
	}

	err := l.client.ListDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("list_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// ListCountInput represents input for eva_list_count tool.
type ListCountInput struct {
	ProjectID  string `json:"project_id,omitempty"`
	StatusType string `json:"status_type,omitempty"`
	Type       string `json:"type,omitempty"` // "sprint" or "release"
}

// ListCount counts lists.
func (l *ListTools) ListCount(ctx context.Context, input *ListCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityList)

	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}
	if input.StatusType != "" {
		qb = qb.Where(sq.Eq{"cache_status_type": input.StatusType})
	}

	switch input.Type {
	case listTypeSprint:
		qb = qb.Where(sq.Like{"code": evateamclient.ListCodePrefixSprint + "%"})
	case listTypeRelease:
		qb = qb.Where(sq.Like{"code": evateamclient.ListCodePrefixRelease + "%"})
	}

	count, err := l.client.ListCount(ctx, qb)
	if err != nil {
		return nil, WrapError("list_count", err)
	}

	return &CountResult{Count: count}, nil
}

// SprintList is an alias for ListList with type=sprint.
func (l *ListTools) SprintList(ctx context.Context, input *ListListInput) (*ListResult, error) {
	input.Type = listTypeSprint
	return l.ListList(ctx, input)
}

// SprintGet is an alias for ListGet (validates sprint prefix).
func (l *ListTools) SprintGet(ctx context.Context, input *ListGetInput) (any, error) {
	return l.ListGet(ctx, input)
}

// ReleaseList is an alias for ListList with type=release.
func (l *ListTools) ReleaseList(ctx context.Context, input *ListListInput) (*ListResult, error) {
	input.Type = listTypeRelease
	return l.ListList(ctx, input)
}

// ReleaseGet is an alias for ListGet (validates release prefix).
func (l *ListTools) ReleaseGet(ctx context.Context, input *ListGetInput) (any, error) {
	return l.ListGet(ctx, input)
}
