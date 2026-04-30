/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/raoptimus/evateamclient.go/models"
)

// Logic type field constants for type-safe queries
const (
	LogicTypeFieldID           = "id"
	LogicTypeFieldClassName    = "class_name"
	LogicTypeFieldName         = "name"
	LogicTypeFieldCode         = "code"
	LogicTypeFieldCmfModelName = "cmf_model_name"
	LogicTypeFieldParentID     = "parent_id"
	LogicTypeFieldProjectID    = "project_id"
)

// Well-known logic type codes for tasks. Actual codes are install-specific
// and may differ between deployments. The values below match the common defaults.
// Note: LogicTypeEpic ("task.epic") in epic.go is the legacy short form
// kept for backward compatibility with older deployments.
const (
	LogicTypeCodeEpic  = "task.epic:default"
	LogicTypeCodeStory = "task.userstory:story"
	LogicTypeCodeTask  = "task.agile:task"
	LogicTypeCodeBug   = "task.bug:default"
)

// DefaultLogicTypeFields - standard projection for LogicType queries
var DefaultLogicTypeFields = []string{
	LogicTypeFieldID,
	LogicTypeFieldClassName,
	LogicTypeFieldName,
	LogicTypeFieldCode,
	LogicTypeFieldCmfModelName,
}

// LogicTypeList retrieves logic types using a QueryBuilder.
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityLogicType).
//	  Where(sq.Eq{"code": "task.epic"})
//	items, meta, err := client.LogicTypeList(ctx, qb)
func (c *Client) LogicTypeList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.LogicType, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultLogicTypeFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfLogicType.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.LogicTypeListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to list logic types")
	}

	return resp.Result, &resp.Meta, nil
}

// LogicTypeByCode retrieves a single LogicType by its code
// (e.g. "task.epic", "task.story"). Returns an error if no
// logic type with that code is found on the server.
// Example:
//
//	lt, err := client.LogicTypeByCode(ctx, evateamclient.LogicTypeEpic)
//	// lt.ID -> "CmfLogicType:..."
func (c *Client) LogicTypeByCode(
	ctx context.Context,
	code string,
) (*models.LogicType, error) {
	qb := NewQueryBuilder().
		Select(DefaultLogicTypeFields...).
		From(EntityLogicType).
		Where(sq.Eq{LogicTypeFieldCode: code}).
		Limit(1)

	items, _, err := c.LogicTypeList(ctx, qb)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, errors.Errorf("logic type with code %q not found", code)
	}

	return &items[0], nil
}
