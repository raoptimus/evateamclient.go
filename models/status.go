/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package models

// Status type constants for cache_status_type field
const (
	// StatusTypeOpen represents an open/active status
	StatusTypeOpen = "OPEN"

	// StatusTypeInProgress represents an in-progress status
	StatusTypeInProgress = "IN_PROGRESS"

	// StatusTypeClosed represents a closed status
	StatusTypeClosed = "CLOSED"
)

// Status represents a CmfStatus workflow sub-status.
type Status struct {
	ID         string  `json:"id"`
	ClassName  string  `json:"class_name"`
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	StatusType string  `json:"status_type"`
	Color      string  `json:"color,omitempty"`
	WorkflowID string  `json:"workflow_id,omitempty"`
	ProjectID  *string `json:"project_id,omitempty"`
	ParentID   *string `json:"parent_id,omitempty"`
	CmfOwnerID string  `json:"cmf_owner_id,omitempty"`
}
