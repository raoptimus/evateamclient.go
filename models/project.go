package models

import "encoding/json"

// CmfProject represents the complete object returned in "result" of CmfProject.get/list.
type CmfProject struct {
	// Core identification fields (always present)
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	Code      string `json:"code"`
	Name      string `json:"name"`

	// Status and workflow fields
	CacheStatusType string  `json:"cache_status_type,omitempty"`
	WorkflowType    *string `json:"workflow_type,omitempty"`
	WorkflowID      string  `json:"workflow_id,omitempty"`

	// Hierarchy and ownership
	ParentID   *string `json:"parent_id,omitempty"`
	ProjectID  string  `json:"project_id,omitempty"`
	CmfOwnerID string  `json:"cmf_owner_id,omitempty"`

	// System flags
	System         bool `json:"system,omitempty"`
	ImportOriginal bool `json:"import_original,omitempty"`

	// Permissions and security
	SlOwnerLock                      bool            `json:"sl_owner_lock,omitempty"`
	PermParentOwnerID                *string         `json:"perm_parent_owner_id,omitempty"`
	PermInheritACLID                 *string         `json:"perm_inherit_acl_id,omitempty"`
	PermEffectiveACLID               *string         `json:"perm_effective_acl_id,omitempty"`
	PermSecurityLevelAllowedIDsCache json.RawMessage `json:"perm_security_level_allowed_ids_cache,omitempty"`

	// Template flag
	IsTemplate bool `json:"is_template,omitempty"`
}

// CmfProjectGetResponse is the complete response structure for CmfProject.get.
type CmfProjectGetResponse struct {
	JSONRPC string     `json:"jsonrpc,omitempty"`
	Result  CmfProject `json:"result,omitempty"`
	Meta    CmfMeta    `json:"meta,omitempty"`
}

// CmfProjectListResponse is the complete response structure for CmfProject.list.
type CmfProjectListResponse struct {
	JSONRPC string       `json:"jsonrpc,omitempty"`
	Result  []CmfProject `json:"result,omitempty"`
	Meta    CmfMeta      `json:"meta,omitempty"`
}
