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
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// treeCall captures one CmfTask.create invocation for assertions.
type treeCall struct {
	name       string
	epic       string
	parentTask string
	logicType  string
}

// newTreeServer builds a TaskTools backed by an httptest server that:
//   - resolves logic type codes to deterministic ids (CmfLogicType:<code>)
//   - assigns each created task an id/code derived from its name
//   - records every create call for inspection.
//
// failNames lets a test force a create to fail for tasks with the given names.
// The returned accessors expose the recorded create calls and the status applied
// to each task by name (via the CmfTask.update workflow transition).
func newTreeServer(t *testing.T, failNames map[string]bool) (*tools.TaskTools, func() []treeCall, func() map[string]string) {
	t.Helper()

	var mu sync.Mutex
	var calls []treeCall
	statuses := map[string]string{}
	// Map created id -> name so CmfTask.get can echo the right task back.
	idToName := map[string]string{}

	idForName := func(name string) string { return "CmfTask:id-" + name }
	codeForName := func(name string) string { return "T-" + name }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		method := r.URL.Query().Get("m")

		var req struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		_ = json.Unmarshal(body, &req)

		write := func(s string) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(s))
		}

		switch method {
		case "CmfLogicType.list":
			code := extractFieldFromFilter(req.Kwargs, "code")
			write(`{"jsonrpc":"2.2","result":[{"id":"CmfLogicType:` + code + `","class_name":"CmfLogicType","code":"` + code + `"}]}`)

		case "CmfTask.create":
			name, _ := req.Kwargs["name"].(string)
			epic, _ := req.Kwargs["epic"].(string)
			parentTask, _ := req.Kwargs["parent_task"].(string)
			logicType, _ := req.Kwargs["logic_type"].(string)

			mu.Lock()
			calls = append(calls, treeCall{name: name, epic: epic, parentTask: parentTask, logicType: logicType})
			if failNames[name] {
				mu.Unlock()
				write(`{"jsonrpc":"2.2","error":{"code":-32000,"message":"create failed for ` + name + `"}}`)
				return
			}
			id := idForName(name)
			idToName[id] = name
			mu.Unlock()
			write(`{"jsonrpc":"2.2","result":"` + id + `"}`)

		case "CmfTask.update":
			id, _ := req.Args[0].(string)
			mu.Lock()
			if status, ok := req.Kwargs["cache_status_type"].(string); ok {
				statuses[idToName[id]] = status
			}
			mu.Unlock()
			write(`{"jsonrpc":"2.2","result":"` + id + `"}`)

		case "CmfTask.get":
			// EpicByID passes the id in kwargs["id"]; other gets use a filter.
			id := extractFieldFromFilter(req.Kwargs, "id")
			if id == "" {
				id, _ = req.Kwargs["id"].(string)
			}
			mu.Lock()
			name := idToName[id]
			mu.Unlock()
			write(`{"jsonrpc":"2.2","result":{"id":"` + id + `","class_name":"CmfTask","code":"` + codeForName(name) + `","name":"` + name + `"}}`)

		default:
			http.Error(w, "unexpected method: "+method, http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  srv.URL,
		APIToken: "test-token",
	})
	require.NoError(t, err)

	getCalls := func() []treeCall {
		mu.Lock()
		defer mu.Unlock()
		out := make([]treeCall, len(calls))
		copy(out, calls)
		return out
	}
	getStatuses := func() map[string]string {
		mu.Lock()
		defer mu.Unlock()
		out := make(map[string]string, len(statuses))
		maps.Copy(out, statuses)
		return out
	}
	return tools.NewTaskTools(client), getCalls, getStatuses
}

// extractFieldFromFilter pulls the value for a given field out of a kwargs filter,
// supporting both the single []any and the [][]any encodings.
func extractFieldFromFilter(kwargs map[string]any, field string) string {
	filter, ok := kwargs["filter"]
	if !ok {
		return ""
	}
	arr, ok := filter.([]any)
	if !ok {
		return ""
	}
	if len(arr) == 3 {
		if f, ok := arr[0].(string); ok && f == field {
			if v, ok := arr[2].(string); ok {
				return v
			}
		}
	}
	for _, item := range arr {
		if inner, ok := item.([]any); ok && len(inner) == 3 {
			if f, ok := inner[0].(string); ok && f == field {
				if v, ok := inner[2].(string); ok {
					return v
				}
			}
		}
	}
	return ""
}

// TestTaskTools_TaskCreateTree_FullHierarchy_Successfully verifies the happy path:
// an epic, two stories and their tasks are created with the correct epic_id and
// parent_task linkage and the proper logic types per level.
func TestTaskTools_TaskCreateTree_FullHierarchy_Successfully(t *testing.T) {
	tt, getCalls, _ := newTreeServer(t, nil)

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Epics: []tools.EpicNode{
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "Epic"},
				Stories: []tools.StoryNode{
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "Story1"},
						Tasks: []tools.TreeTaskInput{
							{Name: "Task1a"},
							{Name: "Task1b"},
						},
					},
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "Story2"},
						Tasks: []tools.TreeTaskInput{
							{Name: "Task2a"},
						},
					},
				},
			},
		},
	}

	raw, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)

	result, ok := raw.(*tools.TaskCreateTreeResult)
	require.True(t, ok)

	epicID := "CmfTask:id-Epic"
	require.Len(t, result.Epics, 1)
	epic := result.Epics[0]
	assert.Equal(t, epicID, epic.ID)
	assert.Equal(t, "T-Epic", epic.Code)
	assert.Equal(t, 6, result.Total) // 1 epic + 2 stories + 3 tasks
	assert.Equal(t, 0, result.Failed)
	require.Len(t, epic.Stories, 2)

	for _, s := range epic.Stories {
		assert.Empty(t, s.Error)
		assert.NotEmpty(t, s.ID)
		assert.Equal(t, epicID, s.EpicID)
		for _, task := range s.Tasks {
			assert.Empty(t, task.Error)
			assert.NotEmpty(t, task.ID)
			// A task's epic_id chains to its story, not the root epic.
			assert.Equal(t, s.ID, task.EpicID)
		}
	}

	// Verify linkage and logic types per create call.
	calls := getCalls()
	byName := map[string]treeCall{}
	for _, c := range calls {
		byName[c.name] = c
	}

	assert.Equal(t, "CmfLogicType:task.epic:default", byName["Epic"].logicType)
	assert.Empty(t, byName["Epic"].epic)

	assert.Equal(t, "CmfLogicType:task.userstory:story", byName["Story1"].logicType)
	assert.Equal(t, epicID, byName["Story1"].epic)
	assert.Empty(t, byName["Story1"].parentTask)

	assert.Equal(t, "CmfLogicType:task.agile:task", byName["Task1a"].logicType)
	assert.Equal(t, "CmfTask:id-Story1", byName["Task1a"].epic) // task's epic_id is its story
	assert.Equal(t, "CmfTask:id-Story1", byName["Task1a"].parentTask)

	assert.Equal(t, "CmfTask:id-Story2", byName["Task2a"].parentTask)
}

// TestTaskTools_TaskCreateTree_MultipleEpics_Successfully verifies that several
// epics are created in one call, each with its own stories linked via epic_id.
func TestTaskTools_TaskCreateTree_MultipleEpics_Successfully(t *testing.T) {
	tt, getCalls, _ := newTreeServer(t, nil)

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Epics: []tools.EpicNode{
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "EpicA"},
				Stories: []tools.StoryNode{
					{TreeTaskInput: tools.TreeTaskInput{Name: "StoryA1"}},
				},
			},
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "EpicB"},
				Stories: []tools.StoryNode{
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "StoryB1"},
						Tasks:         []tools.TreeTaskInput{{Name: "TaskB1a"}},
					},
				},
			},
		},
	}

	raw, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)

	result := raw.(*tools.TaskCreateTreeResult)
	require.Len(t, result.Epics, 2)
	assert.Equal(t, 5, result.Total) // 2 epics + 2 stories + 1 task
	assert.Equal(t, 0, result.Failed)

	byEpicName := map[string]*tools.EpicNodeResult{}
	for _, e := range result.Epics {
		byEpicName[e.Name] = e
	}

	a := byEpicName["EpicA"]
	b := byEpicName["EpicB"]
	require.NotNil(t, a)
	require.NotNil(t, b)
	require.NotEmpty(t, a.ID)
	require.NotEmpty(t, b.ID)
	assert.NotEqual(t, a.ID, b.ID)

	// Each story is linked to its own epic.
	require.Len(t, a.Stories, 1)
	assert.Equal(t, a.ID, a.Stories[0].EpicID)
	require.Len(t, b.Stories, 1)
	assert.Equal(t, b.ID, b.Stories[0].EpicID)
	require.Len(t, b.Stories[0].Tasks, 1)
	// The task's epic_id chains to its story, not the epic.
	assert.Equal(t, b.Stories[0].ID, b.Stories[0].Tasks[0].EpicID)

	// The story create calls reference the correct epic id.
	calls := getCalls()
	byName := map[string]treeCall{}
	for _, c := range calls {
		byName[c.name] = c
	}
	assert.Equal(t, "CmfTask:id-EpicA", byName["StoryA1"].epic)
	assert.Equal(t, "CmfTask:id-EpicB", byName["StoryB1"].epic)
}

// TestTaskTools_TaskCreateTree_EpicFails_SkipsSubtree verifies that a failing
// epic is recorded with an error, its stories are not created, and sibling
// epics still succeed.
func TestTaskTools_TaskCreateTree_EpicFails_SkipsSubtree(t *testing.T) {
	tt, getCalls, _ := newTreeServer(t, map[string]bool{"EpicBad": true})

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Epics: []tools.EpicNode{
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "EpicBad"},
				Stories:       []tools.StoryNode{{TreeTaskInput: tools.TreeTaskInput{Name: "OrphanStory"}}},
			},
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "EpicGood"},
				Stories:       []tools.StoryNode{{TreeTaskInput: tools.TreeTaskInput{Name: "GoodStory"}}},
			},
		},
	}

	raw, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)

	result := raw.(*tools.TaskCreateTreeResult)
	assert.Equal(t, 1, result.Failed)

	var bad, good *tools.EpicNodeResult
	for _, e := range result.Epics {
		if e.Name == "EpicBad" {
			bad = e
		} else {
			good = e
		}
	}
	require.NotNil(t, bad)
	require.NotNil(t, good)
	assert.NotEmpty(t, bad.Error)
	assert.Empty(t, bad.ID)
	assert.Empty(t, good.Error)
	assert.NotEmpty(t, good.ID)

	for _, c := range getCalls() {
		assert.NotEqual(t, "OrphanStory", c.name, "stories of a failed epic must not be created")
	}
}

// TestTaskTools_TaskCreateTree_StoryFails_SkipsItsTasks verifies that a failing
// story is recorded with an error, its tasks are not created, and sibling
// stories still succeed.
func TestTaskTools_TaskCreateTree_StoryFails_SkipsItsTasks(t *testing.T) {
	tt, getCalls, _ := newTreeServer(t, map[string]bool{"StoryBad": true})

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Epics: []tools.EpicNode{
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "Epic"},
				Stories: []tools.StoryNode{
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "StoryBad"},
						Tasks:         []tools.TreeTaskInput{{Name: "OrphanTask"}},
					},
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "StoryGood"},
						Tasks:         []tools.TreeTaskInput{{Name: "GoodTask"}},
					},
				},
			},
		},
	}

	raw, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)

	result := raw.(*tools.TaskCreateTreeResult)
	assert.Equal(t, 1, result.Failed)
	require.Len(t, result.Epics, 1)

	var bad, good *tools.StoryNodeResult
	for _, s := range result.Epics[0].Stories {
		if s.Name == "StoryBad" {
			bad = s
		} else {
			good = s
		}
	}
	require.NotNil(t, bad)
	require.NotNil(t, good)

	assert.NotEmpty(t, bad.Error)
	assert.Empty(t, good.Error)
	assert.NotEmpty(t, good.ID)

	for _, c := range getCalls() {
		assert.NotEqual(t, "OrphanTask", c.name, "tasks of a failed story must not be created")
	}
}

// TestTaskTools_TaskCreateTree_CustomLogicTypeID_UsedDirectly verifies that a
// logic type passed as a CmfLogicType: ID is used as-is without resolving.
func TestTaskTools_TaskCreateTree_CustomLogicTypeID_UsedDirectly(t *testing.T) {
	tt, getCalls, _ := newTreeServer(t, nil)

	input := &tools.TaskCreateTreeInput{
		ProjectID:     "epud",
		EpicLogicType: "CmfLogicType:custom-epic",
		Epics:         []tools.EpicNode{{TreeTaskInput: tools.TreeTaskInput{Name: "Epic"}}},
	}

	_, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)

	for _, c := range getCalls() {
		if c.name == "Epic" {
			assert.Equal(t, "CmfLogicType:custom-epic", c.logicType)
		}
	}
}

// jsonStr marshals v and returns it as a JSON string literal, i.e. the value
// nested as a JSON-encoded string — what an MCP client that stringifies array
// arguments produces.
func jsonStr(t *testing.T, v any) string {
	t.Helper()
	inner, err := json.Marshal(v)
	require.NoError(t, err)
	quoted, err := json.Marshal(string(inner))
	require.NoError(t, err)
	return string(quoted)
}

// TestTaskTools_TaskCreateTree_Status_Applied verifies that a tree-wide default
// status is applied to every node and that a per-node status overrides it, via
// the workflow transition (CmfTask.update with cache_status_type).
func TestTaskTools_TaskCreateTree_Status_Applied(t *testing.T) {
	tt, _, getStatuses := newTreeServer(t, nil)

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Status:    "Backlog", // default for all nodes
		Epics: []tools.EpicNode{
			{
				TreeTaskInput: tools.TreeTaskInput{Name: "Epic"},
				Stories: []tools.StoryNode{
					{
						TreeTaskInput: tools.TreeTaskInput{Name: "Story"},
						Tasks: []tools.TreeTaskInput{
							{Name: "TaskDefault"},
							{Name: "TaskOpen", Status: "OPEN"}, // per-node override
						},
					},
				},
			},
		},
	}

	raw, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)
	result := raw.(*tools.TaskCreateTreeResult)
	assert.Equal(t, 0, result.Failed)

	statuses := getStatuses()
	assert.Equal(t, "Backlog", statuses["Epic"])
	assert.Equal(t, "Backlog", statuses["Story"])
	assert.Equal(t, "Backlog", statuses["TaskDefault"])
	assert.Equal(t, "OPEN", statuses["TaskOpen"], "per-node status overrides the default")
}

// TestTaskTools_TaskCreateTree_NoStatus_NoTransition verifies that without a
// status no workflow transition is performed.
func TestTaskTools_TaskCreateTree_NoStatus_NoTransition(t *testing.T) {
	tt, _, getStatuses := newTreeServer(t, nil)

	input := &tools.TaskCreateTreeInput{
		ProjectID: "epud",
		Epics: []tools.EpicNode{
			{TreeTaskInput: tools.TreeTaskInput{Name: "Epic"}},
		},
	}

	_, err := tt.TaskCreateTree(context.Background(), input)
	require.NoError(t, err)
	assert.Empty(t, getStatuses(), "no status means no CmfTask.update transition")
}

// TestTaskCreateTreeInput_StringifiedArrays_Coerced verifies that epics, stories,
// tasks and the string-array fields are accepted when an MCP client serialises
// them as JSON-encoded strings instead of native arrays, at every nesting level.
func TestTaskCreateTreeInput_StringifiedArrays_Coerced(t *testing.T) {
	// Build the payload inside-out: each array level is embedded as a JSON string.
	tasksStr := jsonStr(t, []map[string]any{{
		"name":      "Task",
		"tags":      json.RawMessage(jsonStr(t, []string{"TAG-1"})), // stringified tags
		"executors": []string{"CmfPerson:1"},                        // native array
	}})
	storiesStr := jsonStr(t, []map[string]any{{
		"name":  "Story",
		"tasks": json.RawMessage(tasksStr),
	}})
	epicsStr := jsonStr(t, []map[string]any{{
		"name":    "Epic",
		"stories": json.RawMessage(storiesStr),
	}})
	payload := `{"project_id":"epud","epics":` + epicsStr + `}`

	var input tools.TaskCreateTreeInput
	require.NoError(t, json.Unmarshal([]byte(payload), &input))

	require.Len(t, input.Epics, 1)
	epic := input.Epics[0]
	assert.Equal(t, "Epic", epic.Name)

	require.Len(t, epic.Stories, 1)
	story := epic.Stories[0]
	assert.Equal(t, "Story", story.Name)

	require.Len(t, story.Tasks, 1)
	task := story.Tasks[0]
	assert.Equal(t, "Task", task.Name)
	assert.Equal(t, tools.TagList{"TAG-1"}, task.Tags)
	assert.Equal(t, tools.StringList{"CmfPerson:1"}, task.Executors)
}

// TestTaskCreateTreeInput_NativeArrays_StillWork verifies the flexible types do
// not break the normal native-array form.
func TestTaskCreateTreeInput_NativeArrays_StillWork(t *testing.T) {
	payload := `{
		"project_id": "epud",
		"epics": [
			{"name": "Epic", "stories": [
				{"name": "Story", "tasks": [
					{"name": "Task", "tags": ["TAG-1"], "lists": ["SPR-1"]}
				]}
			]}
		]
	}`

	var input tools.TaskCreateTreeInput
	require.NoError(t, json.Unmarshal([]byte(payload), &input))

	require.Len(t, input.Epics, 1)
	require.Len(t, input.Epics[0].Stories, 1)
	require.Len(t, input.Epics[0].Stories[0].Tasks, 1)
	task := input.Epics[0].Stories[0].Tasks[0]
	assert.Equal(t, tools.TagList{"TAG-1"}, task.Tags)
	assert.Equal(t, tools.StringList{"SPR-1"}, task.Lists)
}
