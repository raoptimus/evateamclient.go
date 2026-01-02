package models

import "time"

// StatusHistory represents a status change record from CmfStatusHistory.list.
type StatusHistory struct {
	ID            string     `json:"id"`
	ClassName     string     `json:"class_name,omitempty"`
	Code          string     `json:"code,omitempty"`
	Name          *string    `json:"name,omitempty"`
	ParentID      string     `json:"parent_id,omitempty"`      // entity that changed status
	ProjectID     string     `json:"project_id,omitempty"`     // project context
	OldStatus     *string    `json:"old_status,omitempty"`     // previous status value
	NewStatus     *string    `json:"new_status,omitempty"`     // new status value
	OldStatusID   *string    `json:"old_status_id,omitempty"`  // previous status ID
	NewStatusID   *string    `json:"new_status_id,omitempty"`  // new status ID
	CmfOwnerID    string     `json:"cmf_owner_id,omitempty"`   // user who made the change
	CmfCreatedAt  *time.Time `json:"cmf_created_at,omitempty"` // when status changed
	CmfModifiedAt *time.Time `json:"cmf_modified_at,omitempty"`
}

// StatusHistoryResponse for CmfStatusHistory.get.
type StatusHistoryResponse struct {
	JSONRPC string        `json:"jsonrpc,omitempty"`
	Result  StatusHistory `json:"result,omitempty"`
	Meta    Meta          `json:"meta,omitempty"`
}

// StatusHistoryListResponse for CmfStatusHistory.list.
type StatusHistoryListResponse struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	Result  []StatusHistory `json:"result,omitempty"`
	Meta    Meta            `json:"meta,omitempty"`
}
