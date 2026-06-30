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
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSubtaskServer creates an httptest.Server tailored for TaskCreateWithSubtasks tests.
// The handler function receives the parsed request kwargs and method name and returns
// the raw JSON to write as the response body, plus an optional non-2xx HTTP status.
// Returning ("", 0) from handler causes a 400 response so the caller can detect unexpected calls.
func newSubtaskServer(
	t *testing.T,
	handler func(method string, kwargs map[string]any) (responseJSON string, httpStatus int),
) *tools.TaskTools {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		method := r.URL.Query().Get("m")

		var req struct {
			Kwargs map[string]any `json:"kwargs"`
		}
		_ = json.Unmarshal(body, &req)

		responseBody, status := handler(method, req.Kwargs)
		if status == 0 {
			status = http.StatusOK
		}
		if responseBody == "" {
			http.Error(w, "unexpected call: "+method, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(srv.Close)

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  srv.URL,
		APIToken: "test-token",
	})
	require.NoError(t, err)
	return tools.NewTaskTools(client)
}

// TestTaskTools_TaskCreateWithSubtasks_NoChildren_Successfully verifies that when
// Children is empty the method returns only the parent task with Total=0, Failed=0.
func TestTaskTools_TaskCreateWithSubtasks_NoChildren_Successfully(t *testing.T) {
	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			return `{"jsonrpc":"2.2","result":"CmfTask:parent-001"}`, 0
		case "CmfTask.get":
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:parent-001","class_name":"CmfTask","name":"Parent Task"}}`, 0
		}
		return "", 0
	})

	input := &tools.TaskCreateWithSubtasksInput{
		Name:      "Parent Task",
		ProjectID: "CmfProject:proj-001",
		Children:  []tools.ChildTaskInput{},
	}

	raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

	require.NoError(t, err)
	result, ok := raw.(*tools.TaskCreateWithSubtasksResult)
	require.True(t, ok, "result must be *TaskCreateWithSubtasksResult")
	require.NotNil(t, result.Parent)
	assert.Equal(t, "CmfTask:parent-001", result.Parent.ID)
	assert.Equal(t, "Parent Task", result.Parent.Name)
	assert.Empty(t, result.Children)
	assert.Equal(t, 0, result.Total)
	assert.Equal(t, 0, result.Failed)
}

// TestTaskTools_TaskCreateWithSubtasks_AllChildrenSucceed_Successfully verifies that
// when all children are created successfully the result reflects the correct counts
// and all child entries are populated with their task data.
func TestTaskTools_TaskCreateWithSubtasks_AllChildrenSucceed_Successfully(t *testing.T) {
	// Map child names to their server-side IDs so the handler can serve the right task.
	childTaskIDs := map[string]string{
		"Child Alpha": "CmfTask:child-001",
		"Child Beta":  "CmfTask:child-002",
		"Child Gamma": "CmfTask:child-003",
	}

	// Protect the per-name ID look-ups from concurrent requests.
	var mu sync.Mutex

	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			name, _ := kwargs["name"].(string)
			mu.Lock()
			id, isChild := childTaskIDs[name]
			mu.Unlock()
			if isChild {
				return `{"jsonrpc":"2.2","result":"` + id + `"}`, 0
			}
			// parent
			return `{"jsonrpc":"2.2","result":"CmfTask:parent-001"}`, 0

		case "CmfTask.get":
			// The filter contains the id we passed from create.
			// The filter field is a []any or map; the actual id is in kwargs["filter"].
			id := extractIDFromFilter(kwargs)
			if id == "CmfTask:parent-001" {
				return `{"jsonrpc":"2.2","result":{"id":"CmfTask:parent-001","class_name":"CmfTask","name":"Parent Task"}}`, 0
			}
			mu.Lock()
			defer mu.Unlock()
			for name, childID := range childTaskIDs {
				if childID == id {
					return `{"jsonrpc":"2.2","result":{"id":"` + id + `","class_name":"CmfTask","name":"` + name + `"}}`, 0
				}
			}
			return `{"jsonrpc":"2.2","result":{"id":"` + id + `","class_name":"CmfTask","name":"unknown"}}`, 0
		}
		return "", 0
	})

	input := &tools.TaskCreateWithSubtasksInput{
		Name:      "Parent Task",
		ProjectID: "CmfProject:proj-001",
		Children: []tools.ChildTaskInput{
			{Name: "Child Alpha"},
			{Name: "Child Beta"},
			{Name: "Child Gamma"},
		},
	}

	raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

	require.NoError(t, err)
	result, ok := raw.(*tools.TaskCreateWithSubtasksResult)
	require.True(t, ok)
	assert.Equal(t, "CmfTask:parent-001", result.Parent.ID)
	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 0, result.Failed)
	require.Len(t, result.Children, 3)
	for _, child := range result.Children {
		assert.Empty(t, child.Error, "child %q must not have an error", child.Name)
		assert.NotEmpty(t, child.ID, "child %q must have an ID", child.Name)
	}
}

// TestTaskTools_TaskCreateWithSubtasks_ParentCreateFails_Failure verifies that when
// the parent task creation returns an error the method returns a wrapped error and
// no children are created.
func TestTaskTools_TaskCreateWithSubtasks_ParentCreateFails_Failure(t *testing.T) {
	var childCreateCalled atomic.Int32

	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			name, _ := kwargs["name"].(string)
			if name == "Parent Task" {
				return `{"jsonrpc":"2.2","error":{"code":-32000,"message":"server error: cannot create task"}}`, 0
			}
			childCreateCalled.Add(1)
			return `{"jsonrpc":"2.2","result":"CmfTask:should-not-be-reached"}`, 0
		case "CmfTask.get":
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:unused","class_name":"CmfTask","name":"unused"}}`, 0
		}
		return "", 0
	})

	input := &tools.TaskCreateWithSubtasksInput{
		Name:      "Parent Task",
		ProjectID: "CmfProject:proj-001",
		Children: []tools.ChildTaskInput{
			{Name: "Child 1"},
			{Name: "Child 2"},
		},
	}

	raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "task_create_with_subtasks")
	assert.Nil(t, raw)
	assert.Equal(t, int32(0), childCreateCalled.Load(), "no child tasks should be created when parent fails")
}

// TestTaskTools_TaskCreateWithSubtasks_PartialChildFailure_Failure verifies that when
// some child creations fail the Failed counter reflects the number of failures,
// failed children retain their Name from the input, and successful children have
// their full Task data populated.
func TestTaskTools_TaskCreateWithSubtasks_PartialChildFailure_Failure(t *testing.T) {
	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			name, _ := kwargs["name"].(string)
			switch name {
			case "Parent Task":
				return `{"jsonrpc":"2.2","result":"CmfTask:parent-001"}`, 0
			case "Child OK":
				return `{"jsonrpc":"2.2","result":"CmfTask:child-ok-001"}`, 0
			case "Child Fail":
				return `{"jsonrpc":"2.2","error":{"code":-32000,"message":"child creation failed"}}`, 0
			}
			return `{"jsonrpc":"2.2","result":"CmfTask:fallback"}`, 0

		case "CmfTask.get":
			id := extractIDFromFilter(kwargs)
			switch id {
			case "CmfTask:parent-001":
				return `{"jsonrpc":"2.2","result":{"id":"CmfTask:parent-001","class_name":"CmfTask","name":"Parent Task"}}`, 0
			case "CmfTask:child-ok-001":
				return `{"jsonrpc":"2.2","result":{"id":"CmfTask:child-ok-001","class_name":"CmfTask","name":"Child OK"}}`, 0
			}
			return `{"jsonrpc":"2.2","result":{"id":"` + id + `","class_name":"CmfTask","name":"unknown"}}`, 0
		}
		return "", 0
	})

	input := &tools.TaskCreateWithSubtasksInput{
		Name:      "Parent Task",
		ProjectID: "CmfProject:proj-001",
		Children: []tools.ChildTaskInput{
			{Name: "Child OK"},
			{Name: "Child Fail"},
		},
	}

	raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

	require.NoError(t, err)
	result, ok := raw.(*tools.TaskCreateWithSubtasksResult)
	require.True(t, ok)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 1, result.Failed)
	require.Len(t, result.Children, 2)

	// Locate results by the Name field (order is preserved from input).
	okChild := result.Children[0]
	failChild := result.Children[1]

	assert.Empty(t, okChild.Error)
	assert.Equal(t, "CmfTask:child-ok-001", okChild.ID)

	assert.NotEmpty(t, failChild.Error)
	assert.Contains(t, failChild.Error, "child creation failed")
	// Name from input must be preserved even on failure.
	assert.Equal(t, "Child Fail", failChild.Name)
}

// TestTaskTools_TaskCreateWithSubtasks_WorkersClamping_Successfully verifies that
// workers=0 defaults to 3 and workers=99 is capped at 10. The test ensures no
// panic or deadlock occurs with these boundary inputs.
func TestTaskTools_TaskCreateWithSubtasks_WorkersClamping_Successfully(t *testing.T) {
	tests := []struct {
		name    string
		workers int
	}{
		{name: "workers zero defaults to 3", workers: 0},
		{name: "workers negative defaults to 3", workers: -1},
		{name: "workers above max clamped to 10", workers: 99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
				switch method {
				case "CmfTask.create":
					name, _ := kwargs["name"].(string)
					if name == "Parent Task" {
						return `{"jsonrpc":"2.2","result":"CmfTask:parent-001"}`, 0
					}
					return `{"jsonrpc":"2.2","result":"CmfTask:child-001"}`, 0
				case "CmfTask.get":
					return `{"jsonrpc":"2.2","result":{"id":"CmfTask:parent-001","class_name":"CmfTask","name":"Parent Task"}}`, 0
				}
				return "", 0
			})

			// Use 12 children to exercise the clamping: more than the default (3)
			// and more than the max (10).
			children := make([]tools.ChildTaskInput, 12)
			for i := range children {
				children[i] = tools.ChildTaskInput{Name: "Child Task"}
			}

			input := &tools.TaskCreateWithSubtasksInput{
				Name:      "Parent Task",
				ProjectID: "CmfProject:proj-001",
				Workers:   tc.workers,
				Children:  children,
			}

			raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

			require.NoError(t, err)
			result, ok := raw.(*tools.TaskCreateWithSubtasksResult)
			require.True(t, ok)
			assert.Equal(t, 12, result.Total)
		})
	}
}

// TestTaskTools_TaskCreateWithSubtasks_ChildrenInheritParentID_Successfully verifies
// that each child TaskCreate call receives the parent task's ID as ParentTask in kwargs.
func TestTaskTools_TaskCreateWithSubtasks_ChildrenInheritParentID_Successfully(t *testing.T) {
	var mu sync.Mutex
	capturedParentTasks := []string{}

	tt := newSubtaskServer(t, func(method string, kwargs map[string]any) (string, int) {
		switch method {
		case "CmfTask.create":
			name, _ := kwargs["name"].(string)
			if name == "Parent Task" {
				return `{"jsonrpc":"2.2","result":"CmfTask:parent-42"}`, 0
			}
			// Record the parent_task value set by the child create call.
			parentTask, _ := kwargs["parent_task"].(string)
			mu.Lock()
			capturedParentTasks = append(capturedParentTasks, parentTask)
			mu.Unlock()
			return `{"jsonrpc":"2.2","result":"CmfTask:child-001"}`, 0

		case "CmfTask.get":
			id := extractIDFromFilter(kwargs)
			if id == "CmfTask:parent-42" {
				return `{"jsonrpc":"2.2","result":{"id":"CmfTask:parent-42","class_name":"CmfTask","name":"Parent Task"}}`, 0
			}
			return `{"jsonrpc":"2.2","result":{"id":"CmfTask:child-001","class_name":"CmfTask","name":"Child Task"}}`, 0
		}
		return "", 0
	})

	input := &tools.TaskCreateWithSubtasksInput{
		Name:      "Parent Task",
		ProjectID: "CmfProject:proj-001",
		Children: []tools.ChildTaskInput{
			{Name: "Child Task A"},
			{Name: "Child Task B"},
		},
	}

	raw, err := tt.TaskCreateWithSubtasks(context.Background(), input)

	require.NoError(t, err)
	result, ok := raw.(*tools.TaskCreateWithSubtasksResult)
	require.True(t, ok)
	assert.Equal(t, 2, result.Total)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, capturedParentTasks, 2)
	for i, pt := range capturedParentTasks {
		assert.Equal(t, "CmfTask:parent-42", pt,
			"child %d must have parent_task set to the parent task ID", i)
	}
}

// extractIDFromFilter parses the id value from a CmfTask.get kwargs filter.
// The EVA API encodes the filter as either a []any{"id","==","<id>"} or
// a [][]any with one such entry.
func extractIDFromFilter(kwargs map[string]any) string {
	filter, ok := kwargs["filter"]
	if !ok {
		return ""
	}

	// Try []any first (single filter).
	if arr, ok := filter.([]any); ok && len(arr) == 3 {
		if field, ok := arr[0].(string); ok && field == "id" {
			if id, ok := arr[2].(string); ok {
				return id
			}
		}
	}

	// Try [][]any (multiple filters).
	if outer, ok := filter.([]any); ok {
		for _, item := range outer {
			if inner, ok := item.([]any); ok && len(inner) == 3 {
				if field, ok := inner[0].(string); ok && field == "id" {
					if id, ok := inner[2].(string); ok {
						return id
					}
				}
			}
		}
	}

	return ""
}
