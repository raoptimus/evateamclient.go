package models

// Epic represents COMPLETE epic object.
type Epic struct {
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id,omitempty"`
	Status    string `json:"cache_status_type,omitempty"`
}

type EpicListResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  []Epic `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}
