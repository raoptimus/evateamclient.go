package models

import "time"

// SprintStats represents sprint statistics.
type SprintStats struct {
	SprintID             string         `json:"sprint_id,omitempty"`
	TotalTasks           int            `json:"total_tasks,omitempty"`
	CompletedTasks       int            `json:"completed_tasks,omitempty"`
	TotalStoryPoints     int            `json:"total_story_points,omitempty"`
	CompletedStoryPoints int            `json:"completed_story_points,omitempty"`
	TasksByStatus        map[string]int `json:"tasks_by_status,omitempty"`
	TimeSpentByUser      map[string]int `json:"time_spent_by_user,omitempty"`
}

// ProjectStats represents project statistics.
type ProjectStats struct {
	ProjectID     string `json:"project_id,omitempty"`
	TotalTasks    int    `json:"total_tasks,omitempty"`
	OpenTasks     int    `json:"open_tasks,omitempty"`
	ActiveSprints int    `json:"active_sprints,omitempty"`
	TotalUsers    int    `json:"total_users,omitempty"`
}

// TimeSpentTaskEntry represents time spent on a single task by a person.
type TimeSpentTaskEntry struct {
	TaskID    string `json:"task_id"`
	TaskCode  string `json:"task_code"`
	TaskName  string `json:"task_name"`
	TimeSpent int    `json:"time_spent"` // minutes
}

// TimeSpentPersonEntry represents aggregated time spent by a single person.
type TimeSpentPersonEntry struct {
	PersonID   string               `json:"person_id"`
	PersonName string               `json:"person_name"`
	Tasks      []TimeSpentTaskEntry `json:"tasks"`
	TotalTime  int                  `json:"total_time"` // minutes
}

// TimeSpentStats represents aggregated time spent report grouped by person and task.
type TimeSpentStats struct {
	ProjectID      string                 `json:"project_id"`
	DateFrom       string                 `json:"date_from,omitempty"`
	DateTo         string                 `json:"date_to,omitempty"`
	Persons        []TimeSpentPersonEntry `json:"persons"`
	GrandTotalTime int                    `json:"grand_total_time"` // minutes
}

// SprintExecutorKPIEntry represents KPI metrics for a single executor in sprint report.
type SprintExecutorKPIEntry struct {
	PersonID    string   `json:"person_id"`
	PersonName  string   `json:"person_name"`
	ClosedTasks int      `json:"closed_tasks"`
	TaskCodes   []string `json:"task_codes,omitempty"`
}

// SprintExecutorsKPI represents closed tasks KPI grouped by executor for a sprint.
type SprintExecutorsKPI struct {
	ProjectCode      string                   `json:"project_code"`
	SprintCode       string                   `json:"sprint_code"`
	SprintStartDate  time.Time                `json:"sprint_start_date"`
	SprintEndDate    time.Time                `json:"sprint_end_date"`
	BaselineTasks    int                      `json:"baseline_tasks"`
	ExcludedNewTasks int                      `json:"excluded_new_tasks"`
	TotalClosedTasks int                      `json:"total_closed_tasks"`
	UnassignedClosed int                      `json:"unassigned_closed"`
	Executors        []SprintExecutorKPIEntry `json:"executors"`
}

// CountResponse for count queries.
type CountResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  int64  `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}
