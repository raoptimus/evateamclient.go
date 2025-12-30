package models

// Sprint represents COMPLETE sprint object from List.get/list.
type Sprint struct {
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

// SprintListResponse for List.list (sprints).
type SprintListResponse struct {
	JSONRPC string   `json:"jsonrpc,omitempty"`
	Result  []Sprint `json:"result,omitempty"`
	Meta    Meta     `json:"meta,omitempty"`
}
