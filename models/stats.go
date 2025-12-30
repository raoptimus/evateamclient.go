package models

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

// CountResponse for count queries.
type CountResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  int64  `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}
