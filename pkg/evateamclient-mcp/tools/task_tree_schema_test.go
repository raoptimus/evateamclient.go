/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/require"
)

// TestTaskCreateTreeInputSchema_AcceptsArrayAndString verifies that the relaxed
// input schema validates both native arrays and JSON-encoded strings at every
// array level — this is the exact validation the MCP SDK runs before unmarshalling.
func TestTaskCreateTreeInputSchema_AcceptsArrayAndString(t *testing.T) {
	resolved, err := taskCreateTreeInputSchema().Resolve(&jsonschema.ResolveOptions{ValidateDefaults: true})
	require.NoError(t, err)

	native := map[string]any{
		"project_id": "epud",
		"epics": []any{map[string]any{"name": "E", "stories": []any{
			map[string]any{"name": "S", "tasks": []any{map[string]any{"name": "T", "tags": []any{"TAG-1"}}}},
		}}},
	}
	require.NoError(t, resolved.Validate(&native), "native nested arrays must validate")

	stringified := map[string]any{
		"project_id": "epud",
		"epics":      `[{"name":"E","stories":"[{\"name\":\"S\",\"tasks\":\"[{\\\"name\\\":\\\"T\\\"}]\"}]"}]`,
	}
	require.NoError(t, resolved.Validate(&stringified), "stringified arrays must validate")

	tagsAsString := map[string]any{
		"project_id": "epud",
		"epics":      []any{map[string]any{"name": "E", "tags": `["TAG-1"]`}},
	}
	require.NoError(t, resolved.Validate(&tagsAsString), "stringified leaf string-array must validate")
}
