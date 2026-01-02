package models

// Document represents a document in EVA system
type Document struct {
	ID              string `json:"id"`
	ClassName       string `json:"class_name,omitempty"`
	Code            string `json:"code,omitempty"`
	Name            string `json:"name"`
	Text            string `json:"text,omitempty"`
	ProjectID       string `json:"project_id,omitempty"`
	ParentID        string `json:"parent_id,omitempty"`
	CacheStatusType string `json:"cache_status_type,omitempty"`
	CmfCreatedAt    string `json:"cmf_created_at,omitempty"`
	CmfModifiedAt   string `json:"cmf_modified_at,omitempty"`
	CmfOwnerID      string `json:"cmf_owner_id,omitempty"`
	CmfDeleted      bool   `json:"cmf_deleted,omitempty"`
}

// DocumentResponse for Document.get (single document).
type DocumentResponse struct {
	JSONRPC string   `json:"jsonrpc,omitempty"`
	Result  Document `json:"result,omitempty"`
	Meta    Meta     `json:"meta,omitempty"`
}

// DocumentListResponse for Document.list.
type DocumentListResponse struct {
	JSONRPC string     `json:"jsonrpc,omitempty"`
	Result  []Document `json:"result,omitempty"`
	Meta    Meta       `json:"meta,omitempty"`
}
