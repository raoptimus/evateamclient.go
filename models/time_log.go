package models

// CmfTaskTimeLog represents COMPLETE time log entry from CmfTaskTimeLog.list.
type CmfTaskTimeLog struct {
	ID           string `json:"id"`
	ClassName    string `json:"class_name"`
	TaskID       string `json:"task_id,omitempty"` // "CmfTask:UUID"
	UserID       string `json:"user_id,omitempty"` // "CmfPerson:UUID"
	UserName     string `json:"user_name,omitempty"`
	UserLogin    string `json:"user_login,omitempty"`
	MinutesSpent int    `json:"minutes_spent,omitempty"`
	Date         string `json:"date,omitempty"` // "2025-12-27"
	Description  string `json:"description,omitempty"`
	CreatedAt    string `json:"cmf_created_at,omitempty"`
	ModifiedAt   string `json:"cmf_modified_at,omitempty"`
}

// CmfTaskTimeLogListResponse for CmfTaskTimeLog.list.
type CmfTaskTimeLogListResponse struct {
	JSONRPC string           `json:"jsonrpc,omitempty"`
	Result  []CmfTaskTimeLog `json:"result,omitempty"`
	Meta    CmfMeta          `json:"meta,omitempty"`
}
