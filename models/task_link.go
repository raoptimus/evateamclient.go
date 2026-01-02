package models

// TaskLink represents task relationship (CmfRelationOption).
type TaskLink struct {
	ID            string  `json:"id"`
	ClassName     string  `json:"class_name,omitempty"`
	Code          string  `json:"code,omitempty"`
	Name          *string `json:"name,omitempty"`
	ParentID      *string `json:"parent_id,omitempty"`
	ProjectID     *string `json:"project_id,omitempty"`
	CmfOwnerID    string  `json:"cmf_owner_id,omitempty"`
	CmfCreatedAt  string  `json:"cmf_created_at,omitempty"`
	CmfModifiedAt string  `json:"cmf_modified_at,omitempty"`
}

// TaskLinkResponse for TaskLink.get (single link).
type TaskLinkResponse struct {
	JSONRPC string   `json:"jsonrpc,omitempty"`
	Result  TaskLink `json:"result,omitempty"`
	Meta    Meta     `json:"meta,omitempty"`
}

// TaskLinkListResponse for TaskLink.list.
type TaskLinkListResponse struct {
	JSONRPC string     `json:"jsonrpc,omitempty"`
	Result  []TaskLink `json:"result,omitempty"`
	Meta    Meta       `json:"meta,omitempty"`
}
