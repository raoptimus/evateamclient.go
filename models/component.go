/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package models

type Component struct {
	ID              string   `json:"id"`
	ClassName       string   `json:"class_name"`
	Name            string   `json:"name"`
	Code            string   `json:"code"`
	ProjectID       string   `json:"project_id"`
	ParentID        string   `json:"parent_id"`
	CmfOwnerID      string   `json:"cmf_owner_id"`
	WorkflowID      string   `json:"workflow_id"`
	CacheStatusType string   `json:"cache_status_type"`
	Alias           []string `json:"alias"`
}
