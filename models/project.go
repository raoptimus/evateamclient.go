package models

import (
	"encoding/json"
	"time"
)

// Project represents the complete object returned in "result" of Project.get/list.
type Project struct {
	// Core fields (ALWAYS present)
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Text      string `json:"text,omitempty"`

	// CMF timestamps (PRESENT in response)
	CMFLockedAt   *time.Time `json:"cmf_locked_at,omitempty"`
	CMFCreatedAt  time.Time  `json:"cmf_created_at,omitempty"`
	CMFModifiedAt time.Time  `json:"cmf_modified_at,omitempty"`
	CMFViewedAt   time.Time  `json:"cmf_viewed_at,omitempty"`
	CMFDeleted    bool       `json:"cmf_deleted,omitempty"`
	CMFVersion    string     `json:"cmf_version,omitempty"`

	// Existing fields from current struct
	CacheStatusType                  string          `json:"cache_status_type,omitempty"`
	WorkflowType                     *string         `json:"workflow_type,omitempty"`
	WorkflowID                       string          `json:"workflow_id,omitempty"`
	ParentID                         *string         `json:"parent_id,omitempty"`
	CmfOwnerID                       string          `json:"cmf_owner_id,omitempty"`
	System                           bool            `json:"system,omitempty"`
	ImportOriginal                   bool            `json:"import_original,omitempty"`
	SlOwnerLock                      bool            `json:"sl_owner_lock,omitempty"`
	PermParentOwnerID                *string         `json:"perm_parent_owner_id,omitempty"`
	PermInheritACLID                 *string         `json:"perm_inherit_acl_id,omitempty"`
	PermEffectiveACLID               *string         `json:"perm_effective_acl_id,omitempty"`
	PermSecurityLevelAllowedIDsCache json.RawMessage `json:"perm_security_level_allowed_ids_cache,omitempty"`
	IsTemplate                       bool            `json:"is_template,omitempty"`

	// Relations
	Executors  []*Person `json:"executors,omitempty"`
	Assistants []*Person `json:"cmf_owner_assistants,omitempty"`
	Admins     []*Person `json:"cmfprojectadmins,omitempty"`
	Spectators []*Person `json:"spectators,omitempty"`
}

// ProjectGetResponse is the complete response structure for Project.get.
type ProjectGetResponse struct {
	JSONRPC string  `json:"jsonrpc,omitempty"`
	Result  Project `json:"result,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`
}

// ProjectListResponse is the complete response structure for Project.list.
type ProjectListResponse struct {
	JSONRPC string    `json:"jsonrpc,omitempty"`
	Result  []Project `json:"result,omitempty"`
	Meta    Meta      `json:"meta,omitempty"`
}
