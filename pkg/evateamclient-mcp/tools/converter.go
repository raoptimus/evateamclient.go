/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

const defaultPaginationLimit = 100 // Default limit for pagination

// BuildQuery converts QueryInput to evateamclient.QueryBuilder.
func BuildQuery(entity string, input *QueryInput) (*evateamclient.QueryBuilder, error) {
	qb := evateamclient.NewQueryBuilder().From(entity)

	if input == nil {
		return qb, nil
	}

	// Apply fields projection
	if len(input.Fields) > 0 {
		qb = qb.Select(input.Fields...)
	}

	// Apply filters
	for _, f := range input.Filters {
		pred, err := filterToPredicate(f)
		if err != nil {
			return nil, err
		}
		qb = qb.Where(pred)
	}

	// Apply ordering
	if len(input.OrderBy) > 0 {
		qb = qb.OrderBy(input.OrderBy...)
	}

	// Apply pagination
	if input.Limit > 0 {
		qb = qb.Limit(uint64(input.Limit))
	}
	if input.Offset > 0 {
		qb = qb.Offset(uint64(input.Offset))
	}

	// Include archived
	if input.IncludeArchived {
		qb = qb.IncludeArchived()
	}

	return qb, nil
}

// filterToPredicate converts a Filter to Squirrel predicate.
func filterToPredicate(f Filter) (any, error) {
	switch f.Operator {
	case "==", "=":
		return sq.Eq{f.Field: f.Value}, nil
	case "!=", "<>":
		return sq.NotEq{f.Field: f.Value}, nil
	case ">":
		return sq.Gt{f.Field: f.Value}, nil
	case ">=":
		return sq.GtOrEq{f.Field: f.Value}, nil
	case "<":
		return sq.Lt{f.Field: f.Value}, nil
	case "<=":
		return sq.LtOrEq{f.Field: f.Value}, nil
	case "LIKE", "like":
		return sq.Like{f.Field: f.Value}, nil
	case "contains":
		// EVA-specific "contains" operator - pass as raw kwargs
		return nil, fmt.Errorf("'contains' operator requires raw kwargs, use custom filter")
	default:
		return nil, fmt.Errorf("unsupported operator: %s", f.Operator)
	}
}

// BuildKwargs converts QueryInput to raw kwargs map for EVA-specific operations.
func BuildKwargs(input *QueryInput) map[string]any {
	kwargs := make(map[string]any)

	if input == nil {
		return kwargs
	}

	// Apply fields
	if len(input.Fields) > 0 {
		kwargs["fields"] = []string(input.Fields)
	}

	// Apply filters
	if len(input.Filters) > 0 {
		filters := make([][]any, 0, len(input.Filters))
		for _, f := range input.Filters {
			filters = append(filters, []any{f.Field, f.Operator, f.Value})
		}
		if len(filters) == 1 {
			kwargs["filter"] = filters[0]
		} else {
			kwargs["filter"] = filters
		}
	}

	// Apply ordering
	if len(input.OrderBy) > 0 {
		kwargs["order_by"] = []string(input.OrderBy)
	}

	// Apply pagination (EVA uses slice: [start, end])
	if input.Limit > 0 || input.Offset > 0 {
		start := input.Offset
		end := input.Offset + input.Limit
		if input.Limit == 0 {
			end = start + defaultPaginationLimit // default limit
		}
		kwargs["slice"] = []int{start, end}
	}

	// Include archived
	if input.IncludeArchived {
		kwargs["include_archived"] = true
	}

	return kwargs
}

// TagList is a list of tag codes that tolerates both a JSON array
// (["TAG-001"]) and a JSON-encoded string ("[\"TAG-001\"]") on input,
// since some MCP clients serialise array arguments as strings.
type TagList []string

func (tl *TagList) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*tl = coerceTags(v)
	return nil
}

// coerceTags converts an unknown tags value to []string.
// Handles []string, []any (JSON array), and string (JSON-encoded array).
func coerceTags(v any) []string {
	switch t := v.(type) {
	case []string:
		return t
	case []any:
		tags := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				tags = append(tags, s)
			}
		}
		return tags
	case string:
		var tags []string
		if err := json.Unmarshal([]byte(t), &tags); err == nil {
			return tags
		}
	}
	return nil
}

// unmarshalFlexibleSlice unmarshals data into *out, tolerating a JSON-encoded
// string in place of a JSON array — some MCP clients serialise array arguments
// (including nested ones) as strings. An empty string yields a nil slice.
func unmarshalFlexibleSlice[T any](data []byte, out *[]T) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		if strings.TrimSpace(s) == "" {
			*out = nil
			return nil
		}
		return json.Unmarshal([]byte(s), out)
	}
	return json.Unmarshal(data, out)
}

// StringList is a []string that tolerates a JSON-encoded string on input,
// the same way TagList does, for MCP clients that stringify array arguments.
type StringList []string

func (l *StringList) UnmarshalJSON(data []byte) error {
	return unmarshalFlexibleSlice(data, (*[]string)(l))
}

// FilterList is a []Filter that tolerates a JSON-encoded string on input,
// the same way StringList does, for MCP clients that stringify array arguments.
type FilterList []Filter

func (l *FilterList) UnmarshalJSON(data []byte) error {
	return unmarshalFlexibleSlice(data, (*[]Filter)(l))
}

// toAnySlice converts typed slice to []any
func toAnySlice[T any](items []T) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
