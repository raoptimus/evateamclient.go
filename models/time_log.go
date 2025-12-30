package models

import "time"

// TaskTimeLog represents COMPLETE time log entry from TaskTimeLog.list.
type TaskTimeLog struct {
	ID          string     `json:"id"` // "CmfTimeTrackerHistory:xxx"
	Code        string     `json:"code"`
	TimeSpent   float64    `json:"time_spent"` // 1.5 hours
	Description string     `json:"description"`
	Parent      string     `json:"parent"` // "CmfTask:ff513e16-e4c4-11f0-86ab-029a32f97d49"
	Author      *Person    `json:"author,omitempty"`
	CreatedAt   *time.Time `json:"cmf_created_at,omitempty"`
}

// TaskTimeLogListResponse for TaskTimeLog.list.
type TaskTimeLogListResponse struct {
	JSONRPC string        `json:"jsonrpc,omitempty"`
	Result  []TaskTimeLog `json:"result,omitempty"`
	Meta    Meta          `json:"meta,omitempty"`
}
