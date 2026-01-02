package models

import "time"

// TimeLogParent represents nested parent task info in time log response.
type TimeLogParent struct {
	ID         string `json:"id"`
	ClassName  string `json:"class_name"`
	ParentID   string `json:"parent_id"`
	ProjectID  string `json:"project_id"`
	CmfOwnerID string `json:"cmf_owner_id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	WorkflowID string `json:"workflow_id"`
}

// TimeLog represents time log entry from CmfTimeTrackerHistory.list.
type TimeLog struct {
	ID         string         `json:"id"`
	ClassName  string         `json:"class_name"`
	Code       string         `json:"code"`
	Name       *string        `json:"name"`
	TimeSpent  int            `json:"time_spent"` // minutes
	Parent     *TimeLogParent `json:"parent,omitempty"`
	ParentID   string         `json:"parent_id"`
	ProjectID  string         `json:"project_id"`
	CmfOwnerID string         `json:"cmf_owner_id"`
	CreatedAt  *time.Time     `json:"cmf_created_at,omitempty"`
}

// TimeLogResponse for CmfTimeTrackerHistory.get.
type TimeLogResponse struct {
	JSONRPC string  `json:"jsonrpc,omitempty"`
	Result  TimeLog `json:"result,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`
}

// TimeLogListResponse for CmfTimeTrackerHistory.list.
type TimeLogListResponse struct {
	JSONRPC string    `json:"jsonrpc,omitempty"`
	Result  []TimeLog `json:"result,omitempty"`
	Meta    Meta      `json:"meta,omitempty"`
}
