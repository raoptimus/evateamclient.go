package models

// Task represents COMPLETE task object from Task.get/list.
type Task struct {
	// Core identification
	ID               string  `json:"id"`
	ClassName        string  `json:"class_name"`
	Code             string  `json:"code"`
	Name             string  `json:"name"`
	Text             string  `json:"text,omitempty"`
	AgileStoryPoints string  `json:"agile_story_points"` // Story Points
	CacheStatusType  string  `json:"cache_status_type"`
	Priority         int     `json:"priority,omitempty"`
	ProjectID        string  `json:"project_id,omitempty"`
	ParentID         string  `json:"parent_id,omitempty"`
	Deadline         *string `json:"deadline,omitempty"`
	Mark             string  `json:"mark,omitempty"`
	AlarmDate        *string `json:"alarm_date,omitempty"`

	// Nested relations (embedded objects)
	Responsible *Person    `json:"responsible,omitempty"`
	LogicType   *LogicType `json:"logic_type,omitempty"`
	Epic        *Task      `json:"epic,omitempty"` // null or nested task
	WaitingFor  *Person    `json:"waiting_for,omitempty"`
	ParentTask  string     `json:"parent_task,omitempty"`

	// Arrays
	Components  []*Component `json:"components,omitempty"`
	Lists       []*List      `json:"lists,omitempty"` // Sprints
	FixVersions []*List      `json:"fix_versions,omitempty"`
	Tags        []*Tag       `json:"tags,omitempty"`
	Executors   []*Person    `json:"executors,omitempty"`
	Spectators  []*Person    `json:"spectators,omitempty"`

	// System Fields
	CmfLockedAt   *string `json:"cmf_locked_at,omitempty"`
	CmfCreatedAt  string  `json:"cmf_created_at,omitempty"`
	CmfModifiedAt string  `json:"cmf_modified_at,omitempty"`
	CmfViewedAt   *string `json:"cmf_viewed_at,omitempty"`
	CmfDeleted    bool    `json:"cmf_deleted,omitempty"`
	CmfVersion    string  `json:"cmf_version,omitempty"`

	// Import Data
	ImportRawJSON any `json:"import_raw_json,omitempty"`

	// Additional Fields
	ExtID     string `json:"ext_id,omitempty"`
	Approved  bool   `json:"approved,omitempty"`
	IsPublic  bool   `json:"is_public,omitempty"`
	NoControl bool   `json:"no_control,omitempty"`
	IsFlagged bool   `json:"is_flagged,omitempty"`

	// Dates
	PlanStartDate  *string `json:"plan_start_date,omitempty"`
	PlanEndDate    *string `json:"plan_end_date,omitempty"`
	PeriodInterval *string `json:"period_interval,omitempty"`
	PeriodNextDate *string `json:"period_next_date,omitempty"`

	// Status Tracking
	StatusModifiedAt      *string `json:"status_modified_at,omitempty"`
	StatusInProgressStart *string `json:"status_in_progress_start,omitempty"`
	StatusInProgressEnd   *string `json:"status_in_progress_end,omitempty"`
	StatusReviewAt        *string `json:"status_review_at,omitempty"`
	StatusClosedAt        *string `json:"status_closed_at,omitempty"`

	// Additional Flags
	ArchiveDate *string `json:"archiveddate,omitempty"`
	ResultText  string  `json:"result_text,omitempty"`

	// System fields
	CacheChildTasksCount int    `json:"cache_child_tasks_count"`
	WorkflowID           string `json:"workflow_id"`
	CmfOwnerID           string `json:"cmf_owner_id"`
	EpicID               string `json:"epic_id,omitempty"`
	LogicTypeID          string `json:"logic_type_id,omitempty"`
	ResponsibleID        string `json:"responsible_id,omitempty"`
}

// TaskResponse for Task.get (single task).
type TaskResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  Task   `json:"result"`
	Meta    Meta   `json:"meta,omitempty"`
}

// TaskListResponse for Task.list.
type TaskListResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  []Task `json:"result"`
	Meta    Meta   `json:"meta,omitempty"`
}
