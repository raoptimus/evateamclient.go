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

// LogicTypeTools provides MCP tool handlers for logic type operations.
type LogicTypeTools struct {
	client *evateamclient.Client
}

// NewLogicTypeTools creates a new LogicTypeTools instance.
func NewLogicTypeTools(client *evateamclient.Client) *LogicTypeTools {
	return &LogicTypeTools{client: client}
}

// LogicTypeListInput represents input for eva_logic_type_list tool.
type LogicTypeListInput struct {
	QueryInput

	// Optional filter by CMF model name (e.g. "CmfTask").
	CmfModelName string `json:"cmf_model_name,omitempty"`

	// Optional filter by exact code (e.g. "task.epic:default").
	Code string `json:"code,omitempty"`
}

// LogicTypeList returns a list of logic types matching filters.
func (l *LogicTypeTools) LogicTypeList(ctx context.Context, input *LogicTypeListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityLogicType, &input.QueryInput)
	if err != nil {
		return nil, WrapError("logic_type_list", err)
	}

	if input.CmfModelName != "" {
		qb = qb.Where(sq.Eq{evateamclient.LogicTypeFieldCmfModelName: input.CmfModelName})
	}
	if input.Code != "" {
		qb = qb.Where(sq.Eq{evateamclient.LogicTypeFieldCode: input.Code})
	}

	items, _, err := l.client.LogicTypeList(ctx, qb)
	if err != nil {
		return nil, WrapError("logic_type_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(items),
		HasMore: len(items) == input.Limit && input.Limit > 0,
	}, nil
}

// LogicTypeGetInput represents input for eva_logic_type_get tool.
type LogicTypeGetInput struct {
	// LogicType code (e.g. "task.epic:default").
	Code string `json:"code"`
}

// LogicTypeGet retrieves a single logic type by code.
func (l *LogicTypeTools) LogicTypeGet(ctx context.Context, input *LogicTypeGetInput) (any, error) {
	if input.Code == "" {
		return nil, WrapError("logic_type_get", ErrInvalidInput)
	}

	lt, err := l.client.LogicTypeByCode(ctx, input.Code)
	if err != nil {
		return nil, WrapError("logic_type_get", err)
	}

	return lt, nil
}
