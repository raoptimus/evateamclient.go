package models

// CmfTask represents COMPLETE task object from CmfTask.get/list.
type CmfTask struct {
	ID           string   `json:"id"`
	ClassName    string   `json:"class_name"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Text         string   `json:"text,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	SprintIDs    []string `json:"lists,omitempty"` // Sprint codes
	CmfOwnerID   string   `json:"cmf_owner_id,omitempty"`
	Responsible  string   `json:"responsible,omitempty"`
	WaitingFor   string   `json:"waiting_for,omitempty"`
	Executors    []string `json:"executors,omitempty"`
	Spectators   []string `json:"spectators,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	Mark         string   `json:"mark,omitempty"`
	AlarmDate    *string  `json:"alarm_date,omitempty"`
	Deadline     *string  `json:"deadline,omitempty"`
	EpicID       string   `json:"epic,omitempty"`
	ParentTaskID string   `json:"parent_task,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	CacheStatus  string   `json:"cache_status_type,omitempty"`
}

// CmfTaskListResponse for CmfTask.list.
type CmfTaskListResponse struct {
	JSONRPC string    `json:"jsonrpc,omitempty"`
	Result  []CmfTask `json:"result,omitempty"`
	Meta    CmfMeta   `json:"meta,omitempty"`
}
