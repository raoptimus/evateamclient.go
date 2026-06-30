/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskCreate_Tags_ArrayPassthrough verifies that a JSON array of tag codes
// reaches the CmfTask.create kwargs unchanged.
func TestTaskCreate_Tags_ArrayPassthrough(t *testing.T) {
	var captured map[string]any
	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			captured = kwargs
			return `{"jsonrpc":"2.2","result":"CmfTask:test-id"}`, 0
		case "CmfTask.get":
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:test-id","class_name":"CmfTask","name":"Test"}}`, 0
		}
		return "", 0
	})

	var input tools.TaskCreateInput
	require.NoError(t, json.Unmarshal(
		[]byte(`{"name":"Test","project_id":"CmfProject:p","tags":["TAG-001","TAG-002"]}`),
		&input,
	))

	_, err := tt.TaskCreate(context.Background(), &input)

	require.NoError(t, err)
	assert.Equal(t, []any{"TAG-001", "TAG-002"}, captured["tags"])
}

// TestTaskCreate_Tags_JSONEncodedString is the regression test for the bug:
// some MCP clients serialise the tags array as a JSON-encoded string. It must
// still reach the API as a real array.
func TestTaskCreate_Tags_JSONEncodedString(t *testing.T) {
	var captured map[string]any
	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			captured = kwargs
			return `{"jsonrpc":"2.2","result":"CmfTask:test-id"}`, 0
		case "CmfTask.get":
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:test-id","class_name":"CmfTask","name":"Test"}}`, 0
		}
		return "", 0
	})

	var input tools.TaskCreateInput
	require.NoError(t, json.Unmarshal(
		[]byte(`{"name":"Test","project_id":"CmfProject:p","tags":"[\"TAG-004379\",\"TAG-004400\"]"}`),
		&input,
	))

	_, err := tt.TaskCreate(context.Background(), &input)

	require.NoError(t, err)
	assert.Equal(t, []any{"TAG-004379", "TAG-004400"}, captured["tags"])
}

// TestTaskCreate_Tags_Missing verifies that when no tags are supplied the
// CmfTask.create kwargs omit the tags field entirely.
func TestTaskCreate_Tags_Missing(t *testing.T) {
	var captured map[string]any
	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			captured = kwargs
			return `{"jsonrpc":"2.2","result":"CmfTask:test-id"}`, 0
		case "CmfTask.get":
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:test-id","class_name":"CmfTask","name":"Test"}}`, 0
		}
		return "", 0
	})

	var input tools.TaskCreateInput
	require.NoError(t, json.Unmarshal(
		[]byte(`{"name":"Test","project_id":"CmfProject:p"}`),
		&input,
	))

	_, err := tt.TaskCreate(context.Background(), &input)

	require.NoError(t, err)
	assert.Nil(t, captured["tags"])
}

// TestTaskCreateWithSubtasks_Tags_JSONEncodedString verifies that tags survive
// the JSON-encoded-string form on both the parent and a child task.
func TestTaskCreateWithSubtasks_Tags_JSONEncodedString(t *testing.T) {
	var mu sync.Mutex
	capturedTags := map[string]any{} // task name -> tags kwargs

	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			name, _ := kwargs["name"].(string)
			mu.Lock()
			capturedTags[name] = kwargs["tags"]
			mu.Unlock()
			if name == "Parent Task" {
				return `{"jsonrpc":"2.2","result":"CmfTask:parent-001"}`, 0
			}
			return `{"jsonrpc":"2.2","result":"CmfTask:child-001"}`, 0
		case "CmfTask.get":
			id := extractIDFromFilter(kwargs)
			return `{"jsonrpc":"2.2","result":{"id":"` + id + `","class_name":"CmfTask","name":"x"}}`, 0
		}
		return "", 0
	})

	var input tools.TaskCreateWithSubtasksInput
	require.NoError(t, json.Unmarshal(
		[]byte(`{"name":"Parent Task","project_id":"CmfProject:p",`+
			`"tags":"[\"TAG-PARENT\"]",`+
			`"children":[{"name":"Child A","tags":"[\"TAG-CHILD\"]"}]}`),
		&input,
	))

	_, err := tt.TaskCreateWithSubtasks(context.Background(), &input)

	require.NoError(t, err)
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []any{"TAG-PARENT"}, capturedTags["Parent Task"])
	assert.Equal(t, []any{"TAG-CHILD"}, capturedTags["Child A"])
}
