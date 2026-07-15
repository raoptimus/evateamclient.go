/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"encoding/json"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateAgainst resolves the relaxed input schema for In and validates value
// against it — this is exactly the check the MCP SDK runs before unmarshalling.
func validateAgainst[In any](t *testing.T, value map[string]any) error {
	t.Helper()
	resolved, err := relaxedInputSchema[In]().Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	require.NoError(t, err)
	return resolved.Validate(&value)
}

// TestRelaxedSchema_TaskCreate_AcceptsArrayAndString pins report A1: the tags
// array must validate both as a native array and as a JSON-encoded string.
func TestRelaxedSchema_TaskCreate_AcceptsArrayAndString(t *testing.T) {
	require.NoError(t, validateAgainst[TaskCreateInput](t, map[string]any{
		"name": "T", "project_id": "epud", "tags": []any{"TAG-1"},
	}), "native tags array must validate")

	require.NoError(t, validateAgainst[TaskCreateInput](t, map[string]any{
		"name": "T", "project_id": "epud", "tags": `["TAG-1"]`,
	}), "stringified tags array must validate")

	require.NoError(t, validateAgainst[TaskCreateInput](t, map[string]any{
		"name": "T", "project_id": "epud", "executors": `["Person:1"]`, "lists": `["List:1"]`,
	}), "stringified executors/lists must validate")
}

// TestRelaxedSchema_QueryInput_AcceptsArrayAndString pins report A3: fields,
// filters and order_by must validate both as native arrays and as strings, for
// every list tool that embeds QueryInput (task_list, tag_list, ...).
func TestRelaxedSchema_QueryInput_AcceptsArrayAndString(t *testing.T) {
	native := map[string]any{
		"fields":   []any{"code", "name"},
		"filters":  []any{map[string]any{"field": "priority", "operator": ">", "value": 1}},
		"order_by": []any{"-priority"},
	}
	require.NoError(t, validateAgainst[TaskListInput](t, native), "native arrays must validate (task_list)")
	require.NoError(t, validateAgainst[TagListInput](t, native), "native arrays must validate (tag_list)")

	stringified := map[string]any{
		"fields":   `["code","name"]`,
		"filters":  `[{"field":"priority","operator":">","value":1}]`,
		"order_by": `["-priority"]`,
	}
	require.NoError(t, validateAgainst[TaskListInput](t, stringified), "stringified arrays must validate (task_list)")
	require.NoError(t, validateAgainst[TagListInput](t, stringified), "stringified arrays must validate (tag_list)")
}

// TestQueryInput_UnmarshalCoercesStringifiedArrays pins the coercion layer:
// once the relaxed schema lets a stringified array through, UnmarshalJSON must
// decode it into a real slice.
func TestQueryInput_UnmarshalCoercesStringifiedArrays(t *testing.T) {
	raw := []byte(`{
		"fields": "[\"code\",\"name\"]",
		"filters": "[{\"field\":\"priority\",\"operator\":\">\",\"value\":1}]",
		"order_by": "[\"-priority\"]"
	}`)

	var in TaskListInput
	require.NoError(t, json.Unmarshal(raw, &in))

	assert.Equal(t, StringList{"code", "name"}, in.Fields)
	assert.Equal(t, StringList{"-priority"}, in.OrderBy)
	require.Len(t, in.Filters, 1)
	assert.Equal(t, "priority", in.Filters[0].Field)
	assert.Equal(t, ">", in.Filters[0].Operator)
}

// TestTaskCreateInput_UnmarshalCoercesStringifiedArrays pins the coercion layer
// for eva_task_create executors/lists (tags are already covered elsewhere).
func TestTaskCreateInput_UnmarshalCoercesStringifiedArrays(t *testing.T) {
	raw := []byte(`{
		"name": "T",
		"project_id": "epud",
		"tags": "[\"TAG-1\"]",
		"executors": "[\"Person:1\"]",
		"lists": "[\"List:1\"]"
	}`)

	var in TaskCreateInput
	require.NoError(t, json.Unmarshal(raw, &in))

	assert.Equal(t, TagList{"TAG-1"}, in.Tags)
	assert.Equal(t, StringList{"Person:1"}, in.Executors)
	assert.Equal(t, StringList{"List:1"}, in.Lists)
}
