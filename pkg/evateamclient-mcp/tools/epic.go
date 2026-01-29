package tools

import (
	"context"

	evateamclient "github.com/raoptimus/evateamclient"
)

// EpicTools provides MCP tool handlers for epic operations.
type EpicTools struct {
	client *evateamclient.Client
}

// NewEpicTools creates a new EpicTools instance.
func NewEpicTools(client *evateamclient.Client) *EpicTools {
	return &EpicTools{client: client}
}

// EpicListInput represents input for eva_epic_list tool.
type EpicListInput struct {
	QueryInput
	ProjectID string `json:"project_id,omitempty"`
}

// EpicList returns a list of epics.
func (e *EpicTools) EpicList(ctx context.Context, input EpicListInput) (*ListResult, error) {
	// Build kwargs with logic_type.code filter
	kwargs := BuildKwargs(&input.QueryInput)

	var filters [][]any
	if existingFilter, ok := kwargs["filter"].([][]any); ok {
		filters = existingFilter
	} else if singleFilter, ok := kwargs["filter"].([]any); ok {
		filters = [][]any{singleFilter}
	}

	// Add epic filter
	filters = append(filters, []any{"logic_type.code", "==", evateamclient.LogicTypeEpic})

	if input.ProjectID != "" {
		filters = append(filters, []any{"project_id", "==", input.ProjectID})
	}

	kwargs["filter"] = filters

	// Set default fields if not specified
	if _, ok := kwargs["fields"]; !ok {
		kwargs["fields"] = evateamclient.DefaultEpicListFields
	}

	epics, _, err := e.client.Epics(ctx, kwargs)
	if err != nil {
		return nil, WrapError("epic_list", err)
	}

	return &ListResult{
		Items:   epics,
		HasMore: len(epics) == input.Limit && input.Limit > 0,
	}, nil
}

// EpicGetInput represents input for eva_epic_get tool.
type EpicGetInput struct {
	Code   string   `json:"code,omitempty"`
	ID     string   `json:"id,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

// EpicGet retrieves a single epic.
func (e *EpicTools) EpicGet(ctx context.Context, input EpicGetInput) (any, error) {
	switch {
	case input.Code != "":
		epic, _, err := e.client.Epic(ctx, input.Code, input.Fields)
		if err != nil {
			return nil, WrapError("epic_get", err)
		}
		return epic, nil
	case input.ID != "":
		epic, _, err := e.client.EpicByID(ctx, input.ID, input.Fields)
		if err != nil {
			return nil, WrapError("epic_get", err)
		}
		return epic, nil
	default:
		return nil, WrapError("epic_get", ErrInvalidInput)
	}
}

// EpicCountInput represents input for eva_epic_count tool.
type EpicCountInput struct {
	ProjectID string `json:"project_id,omitempty"`
}

// EpicCount counts epics.
func (e *EpicTools) EpicCount(ctx context.Context, input EpicCountInput) (*CountResult, error) {
	// Build kwargs with logic_type.code filter
	kwargs := make(map[string]any)

	filters := [][]any{
		{"logic_type.code", "==", evateamclient.LogicTypeEpic},
	}

	if input.ProjectID != "" {
		filters = append(filters, []any{"project_id", "==", input.ProjectID})
	}

	kwargs["filter"] = filters

	count, _, err := e.client.TasksCount(ctx, kwargs)
	if err != nil {
		return nil, WrapError("epic_count", err)
	}

	return &CountResult{Count: int(count)}, nil
}
