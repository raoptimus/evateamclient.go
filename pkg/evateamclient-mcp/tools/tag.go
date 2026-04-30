/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

// TagTools provides MCP tool handlers for tag operations.
type TagTools struct {
	client *evateamclient.Client
}

// NewTagTools creates a new TagTools instance.
func NewTagTools(client *evateamclient.Client) *TagTools {
	return &TagTools{client: client}
}

// TagListInput represents input for eva_tag_list tool.
type TagListInput struct {
	QueryInput

	// Optional filter by project ID or code (e.g. "epud").
	ProjectID string `json:"project_id,omitempty"`

	// Optional filter by tag name (exact match).
	Name string `json:"name,omitempty"`
}

// TagList returns a list of tags matching filters.
func (t *TagTools) TagList(ctx context.Context, input *TagListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityTag, &input.QueryInput)
	if err != nil {
		return nil, WrapError("tag_list", err)
	}

	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{evateamclient.TagFieldProjectID: input.ProjectID})
	}
	if input.Name != "" {
		qb = qb.Where(sq.Eq{evateamclient.TagFieldName: input.Name})
	}

	items, _, err := t.client.TagList(ctx, qb)
	if err != nil {
		return nil, WrapError("tag_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(items),
		HasMore: len(items) == input.Limit && input.Limit > 0,
	}, nil
}
