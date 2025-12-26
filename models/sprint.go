package models

// CmfSprint represents COMPLETE sprint object from CmfList.get/list.
type CmfSprint struct {
	ID          string  `json:"id"`
	ClassName   string  `json:"class_name"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Parent      string  `json:"parent,omitempty"` // "code_project"
	CacheStatus string  `json:"cache_status_type,omitempty"`
	WorkflowID  string  `json:"workflow_id,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
	Goal        string  `json:"goal,omitempty"`
	System      bool    `json:"system,omitempty"`
	SlOwnerLock bool    `json:"sl_owner_lock,omitempty"`
}

// CmfSprintListResponse for CmfList.list (sprints).
type CmfSprintListResponse struct {
	JSONRPC string      `json:"jsonrpc,omitempty"`
	Result  []CmfSprint `json:"result,omitempty"`
	Meta    CmfMeta     `json:"meta,omitempty"`
}
