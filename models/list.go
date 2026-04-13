/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package models

import (
	"strings"
	"time"
)

const (
	ListSprintPrefix  = "SPR-"
	ListReleasePrefix = "REL-"
)

// ListParent represents nested parent project info in list response.
type ListParent struct {
	ID         string  `json:"id"`
	ClassName  string  `json:"class_name"`
	ParentID   *string `json:"parent_id"`
	ProjectID  string  `json:"project_id"`
	CmfOwnerID string  `json:"cmf_owner_id"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	WorkflowID string  `json:"workflow_id"`
}

// List represents CmfList object (sprint or release) from CmfList.get/list.
// Use IsSprint() or IsRelease() to determine the type by code prefix.
type List struct {
	ID                string      `json:"id"`
	ClassName         string      `json:"class_name"`
	Code              string      `json:"code"`
	Name              string      `json:"name"`
	CacheStatusType   string      `json:"cache_status_type,omitempty"`
	CacheMembersCount int         `json:"cache_members_count,omitempty"`
	LimitDays         string      `json:"limit_days,omitempty"`
	Parent            *ListParent `json:"parent,omitempty"`
	ParentID          string      `json:"parent_id,omitempty"`
	ProjectID         string      `json:"project_id,omitempty"`
	CmfOwnerID        string      `json:"cmf_owner_id,omitempty"`
	WorkflowID        string      `json:"workflow_id,omitempty"`
	PlanStartDate     time.Time   `json:"plan_start_date,omitempty"`
	PlanEndDate       time.Time   `json:"plan_end_date,omitempty"`
	Goal              string      `json:"goal,omitempty"`
	Text              string      `json:"text,omitempty"`
	System            bool        `json:"system,omitempty"`
	SlOwnerLock       bool        `json:"sl_owner_lock,omitempty"`
}

// IsSprint returns true if this list is a sprint (code starts with "SPR-").
func (l *List) IsSprint() bool {
	return strings.HasPrefix(l.Code, ListSprintPrefix)
}

// IsRelease returns true if this list is a release (code starts with "REL-").
func (l *List) IsRelease() bool {
	return strings.HasPrefix(l.Code, ListReleasePrefix)
}

// ListResponse for CmfList.get (single list).
type ListResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  List   `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}

// ListListResponse for CmfList.list.
type ListListResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  []List `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}
