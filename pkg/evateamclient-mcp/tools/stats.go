package tools

import (
	"context"
	"time"

	"github.com/raoptimus/evateamclient.go"
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

// TimeSpentStatsInput represents input for eva_stats_timespent tool.
type TimeSpentStatsInput struct {
	ProjectID string `json:"project_id"`
	DateFrom  string `json:"date_from,omitempty"`
	DateTo    string `json:"date_to,omitempty"`
}

// TimeSpentStats retrieves time spent report grouped by person and task.
func (s *StatsTools) TimeSpentStats(ctx context.Context, input TimeSpentStatsInput) (any, error) {
	if input.ProjectID == "" {
		return nil, WrapError("stats_timespent", ErrInvalidInput)
	}

	stats, err := s.client.TimeSpentStats(ctx, evateamclient.TimeSpentStatsParams{
		ProjectID: input.ProjectID,
		DateFrom:  input.DateFrom,
		DateTo:    input.DateTo,
	})
	if err != nil {
		return nil, WrapError("stats_timespent", err)
	}

	return stats, nil
}

// SprintExecutorsKPIInput represents input for eva_stats_sprint_executors_kpi tool.
type SprintExecutorsKPIInput struct {
	ProjectCode     string    `json:"project_code,omitempty"`
	SprintCode      string    `json:"sprint_code,omitempty"`
	SprintStartDate time.Time `json:"sprint_start_date,omitempty"`
	SprintEndDate   time.Time `json:"sprint_end_date,omitempty"`
}

// SprintExecutorsKPI retrieves KPI report for closed sprint tasks grouped by executor.
// If sprint_code is empty, the report is aggregated across all project sprints.
func (s *StatsTools) SprintExecutorsKPI(ctx context.Context, input *SprintExecutorsKPIInput) (any, error) {
	if input.ProjectCode == "" {
		return nil, WrapError("stats_sprint_executors_kpi", ErrInvalidInput)
	}

	report, err := s.client.SprintExecutorsKPI(ctx, &evateamclient.SprintExecutorsKPIParams{
		SprintCode:      input.SprintCode,
		ProjectCode:     input.ProjectCode,
		SprintStartDate: input.SprintStartDate,
		SprintEndDate:   input.SprintEndDate,
	})
	if err != nil {
		return nil, WrapError("stats_sprint_executors_kpi", err)
	}

	return report, nil
}
