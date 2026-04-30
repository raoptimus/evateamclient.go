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

	"github.com/pkg/errors"
	"github.com/raoptimus/evateamclient.go/models"
)

// Tag field constants for type-safe queries
const (
	TagFieldID        = "id"
	TagFieldClassName = "class_name"
	TagFieldName      = "name"
	TagFieldCode      = "code"
	TagFieldAlias     = "alias"
	TagFieldParentID  = "parent_id"
	TagFieldProjectID = "project_id"
)

// DefaultTagFields - standard projection for Tag queries
var DefaultTagFields = []string{
	TagFieldID,
	TagFieldClassName,
	TagFieldName,
	TagFieldCode,
	TagFieldAlias,
}

// TagList retrieves tags using a QueryBuilder.
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name", "alias").
//	  From(evateamclient.EntityTag).
//	  Where(sq.Eq{"project_id": "CmfProject:uuid"})
//	items, meta, err := client.TagList(ctx, qb)
func (c *Client) TagList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Tag, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTagFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTag.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TagListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, errors.WithMessage(err, "failed to list tags")
	}

	return resp.Result, &resp.Meta, nil
}
