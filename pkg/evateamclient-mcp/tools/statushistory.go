package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	evateamclient "github.com/raoptimus/evateamclient"
)

// StatusHistoryTools provides MCP tool handlers for status history operations.
type StatusHistoryTools struct {
	client *evateamclient.Client
}

// NewStatusHistoryTools creates a new StatusHistoryTools instance.
func NewStatusHistoryTools(client *evateamclient.Client) *StatusHistoryTools {
	return &StatusHistoryTools{client: client}
}

// StatusHistoryListInput represents input for eva_statushistory_list tool.
type StatusHistoryListInput struct {
	QueryInput
	TaskID    string `json:"task_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// StatusHistoryList returns a list of status history entries.
func (s *StatusHistoryTools) StatusHistoryList(ctx context.Context, input StatusHistoryListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityStatusHistory, &input.QueryInput)
	if err != nil {
		return nil, WrapError("statushistory_list", err)
	}

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"parent_id": input.TaskID})
	}
	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}

	// Default order by creation time descending
	if len(input.OrderBy) == 0 {
		qb = qb.OrderBy("-cmf_created_at")
	}

	histories, _, err := s.client.StatusHistoryList(ctx, qb)
	if err != nil {
		return nil, WrapError("statushistory_list", err)
	}

	return &ListResult{
		Items:   histories,
		HasMore: len(histories) == input.Limit && input.Limit > 0,
	}, nil
}

// StatusHistoryGetInput represents input for eva_statushistory_get tool.
type StatusHistoryGetInput struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields,omitempty"`
}

// StatusHistoryGet retrieves a single status history entry.
func (s *StatusHistoryTools) StatusHistoryGet(ctx context.Context, input StatusHistoryGetInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("statushistory_get", ErrInvalidInput)
	}

	history, _, err := s.client.StatusHistory(ctx, input.ID, input.Fields)
	if err != nil {
		return nil, WrapError("statushistory_get", err)
	}

	return history, nil
}

// StatusHistoryCountInput represents input for eva_statushistory_count tool.
type StatusHistoryCountInput struct {
	TaskID    string `json:"task_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// StatusHistoryCount counts status history entries.
func (s *StatusHistoryTools) StatusHistoryCount(ctx context.Context, input StatusHistoryCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityStatusHistory)

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"parent_id": input.TaskID})
	}
	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}

	count, err := s.client.StatusHistoryCount(ctx, qb)
	if err != nil {
		return nil, WrapError("statushistory_count", err)
	}

	return &CountResult{Count: count}, nil
}
