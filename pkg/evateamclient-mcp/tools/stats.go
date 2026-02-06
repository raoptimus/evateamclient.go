package tools

import (
	"context"

	evateamclient "github.com/raoptimus/evateamclient.go"
)

// StatsTools provides MCP tool handlers for statistics operations.
type StatsTools struct {
	client *evateamclient.Client
}

// NewStatsTools creates a new StatsTools instance.
func NewStatsTools(client *evateamclient.Client) *StatsTools {
	return &StatsTools{client: client}
}

// ProjectStatsInput represents input for eva_stats_project tool.
type ProjectStatsInput struct {
	ProjectID string `json:"project_id"`
}

// ProjectStats retrieves project statistics.
func (s *StatsTools) ProjectStats(ctx context.Context, input ProjectStatsInput) (any, error) {
	if input.ProjectID == "" {
		return nil, WrapError("stats_project", ErrInvalidInput)
	}

	stats, _, err := s.client.ProjectStats(ctx, input.ProjectID)
	if err != nil {
		return nil, WrapError("stats_project", err)
	}

	return stats, nil
}

// SprintStatsInput represents input for eva_stats_sprint tool.
type SprintStatsInput struct {
	SprintCode string `json:"sprint_code"`
}

// SprintStats retrieves sprint statistics.
func (s *StatsTools) SprintStats(ctx context.Context, input SprintStatsInput) (any, error) {
	if input.SprintCode == "" {
		return nil, WrapError("stats_sprint", ErrInvalidInput)
	}

	stats, err := s.client.SprintStats(ctx, input.SprintCode)
	if err != nil {
		return nil, WrapError("stats_sprint", err)
	}

	return stats, nil
}
