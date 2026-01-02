package models

// Tag represents a tag/label in EVA system
type Tag struct {
	ID         string   `json:"id"`
	ClassName  string   `json:"class_name"`
	Name       string   `json:"name"`  // "Backend"
	Code       string   `json:"code"`  // "TAG-000005"
	Alias      []string `json:"alias"` // ["Backend"] or []
	ParentID   *string  `json:"parent_id,omitempty"`
	ProjectID  *string  `json:"project_id,omitempty"`
	CmfOwnerID string   `json:"cmf_owner_id"`
}
