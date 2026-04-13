package models

import (
	"strings"
	"time"
)

type TaskBrowse struct {
	ID                   string    `json:"id"`
	ClassName            string    `json:"class_name"`
	AgileStoryPoints     string    `json:"agile_story_points"` // Story Points
	CacheStatusType      string    `json:"cache_status_type"`
	Code                 string    `json:"code"`
	Deadline             time.Time `json:"deadline,omitempty"`
	EpicID               string    `json:"epic_id,omitempty"`
	Name                 string    `json:"name"`
	Priority             int       `json:"priority,omitempty"`
	ProjectID            string    `json:"project_id,omitempty"`
	ResponsibleID        string    `json:"responsible_id,omitempty"`
	CacheChildTasksCount int       `json:"cache_child_tasks_count"`
	ParentID             string    `json:"parent_id,omitempty"` // ParentID - project id
	CmfOwnerID           string    `json:"cmf_owner_id"`
	WorkflowID           string    `json:"workflow_id"`

	StatusClosedAt time.Time `json:"status_closed_at,omitempty"`
}

// Task represents COMPLETE task object from Task.get/list.
type Task struct {
	TaskBrowse
	Text      string  `json:"text,omitempty"`
	Mark      string  `json:"mark,omitempty"`
	AlarmDate *string `json:"alarm_date,omitempty"`

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
	CmfLockedAt   time.Time `json:"cmf_locked_at,omitempty"`
	CmfCreatedAt  time.Time `json:"cmf_created_at,omitempty"`
	CmfModifiedAt time.Time `json:"cmf_modified_at,omitempty"`
	CmfViewedAt   time.Time `json:"cmf_viewed_at,omitempty"`
	CmfDeleted    bool      `json:"cmf_deleted,omitempty"`
	CmfVersion    string    `json:"cmf_version,omitempty"`

	// Import Data
	// ImportRawJSON any `json:"import_raw_json,omitempty"`

	// Additional Fields
	ExtID     string `json:"ext_id,omitempty"`
	Approved  bool   `json:"approved,omitempty"`
	IsPublic  bool   `json:"is_public,omitempty"`
	NoControl bool   `json:"no_control,omitempty"`
	IsFlagged bool   `json:"is_flagged,omitempty"`

	// Dates
	PlanStartDate  time.Time `json:"plan_start_date,omitempty"`
	PlanEndDate    time.Time `json:"plan_end_date,omitempty"`
	PeriodInterval string    `json:"period_interval,omitempty"`
	PeriodNextDate string    `json:"period_next_date,omitempty"`

	// Status Tracking
	StatusModifiedAt      time.Time `json:"status_modified_at,omitempty"`
	StatusInProgressStart time.Time `json:"status_in_progress_start,omitempty"`
	StatusInProgressEnd   time.Time `json:"status_in_progress_end,omitempty"`
	StatusReviewAt        time.Time `json:"status_review_at,omitempty"`

	// Additional Flags
	ArchiveDate time.Time `json:"archiveddate,omitempty"`
	ResultText  string    `json:"result_text,omitempty"`

	// System fields
	LogicTypeID string `json:"logic_type_id,omitempty"`
}

func (t TaskBrowse) IsClosedBetween(since, till time.Time) bool {
	if !strings.EqualFold(t.CacheStatusType, StatusTypeClosed) || t.StatusClosedAt.IsZero() {
		return false
	}

	return t.StatusClosedAt.After(since) && t.StatusClosedAt.Before(till)
}

// TaskResponse for Task.get (single task).
type TaskResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  Task   `json:"result"`
	Meta    Meta   `json:"meta,omitempty"`
}

// TaskListResponse for Task.list.
type TaskListResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	Result  []TaskBrowse `json:"result"`
	Meta    Meta         `json:"meta,omitempty"`
}
