/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package models

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// TaskLink represents task relationship (CmfRelationOption).
type TaskLink struct {
	ID            string           `json:"id"`
	ClassName     string           `json:"class_name,omitempty"`
	Code          string           `json:"code,omitempty"`
	Name          *string          `json:"name,omitempty"`
	ParentID      *string          `json:"parent_id,omitempty"`
	ProjectID     *string          `json:"project_id,omitempty"`
	RelationType  RelationTypeCode `json:"relation_type,omitempty"`
	InLink        *TaskRef         `json:"in_link,omitempty"`
	OutLink       *TaskRef         `json:"out_link,omitempty"`
	CmfOwnerID    string           `json:"cmf_owner_id,omitempty"`
	CmfCreatedAt  time.Time        `json:"cmf_created_at,omitempty"`
	CmfModifiedAt time.Time        `json:"cmf_modified_at,omitempty"`
}

// TaskRef references a task on one side of a TaskLink. The EVA server sends
// this as a bare ID string under a plain field projection, or as a nested
// object (with code/name) when the field is requested via "**" — see
// KB-000325 (relation['in_link']['code']). UnmarshalJSON accepts both forms.
type TaskRef struct {
	ID   string `json:"id"`
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

func (r *TaskRef) UnmarshalJSON(data []byte) error {
	var id string
	if err := json.Unmarshal(data, &id); err == nil {
		r.ID = id
		return nil
	}

	type taskRefAlias TaskRef
	var alias taskRefAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return errors.WithMessage(err, "unmarshal TaskRef")
	}
	*r = TaskRef(alias)
	return nil
}

// RelationTypeCode is a relation_type value. Under a plain field projection
// the EVA server sends it as a bare code string, but per
// doc/dopolnitel_nye_optsii_api_zaprosov_doc-000695.pdf, requesting fields
// via "**" (as in_link/out_link do, KB-000325) can return the relation-type
// object instead. UnmarshalJSON accepts both, always yielding the code as a
// plain string (via its "code" field, falling back to "id"), so callers and
// JSON output keep consuming/emitting a plain string either way.
type RelationTypeCode string

func (c *RelationTypeCode) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*c = RelationTypeCode(s)
		return nil
	}

	var obj struct {
		Code string `json:"code"`
		ID   string `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return errors.WithMessage(err, "unmarshal RelationTypeCode")
	}
	if obj.Code != "" {
		*c = RelationTypeCode(obj.Code)
	} else {
		*c = RelationTypeCode(obj.ID)
	}
	return nil
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
