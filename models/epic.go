package models

// CmfEpic represents COMPLETE epic object.
type CmfEpic struct {
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id,omitempty"`
	Status    string `json:"cache_status_type,omitempty"`
}
