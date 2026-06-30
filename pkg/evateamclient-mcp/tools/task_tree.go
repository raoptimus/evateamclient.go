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
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/models"
)

const (
	defaultTreeWorkers = 3
	maxTreeWorkers     = 10
)

// TreeTaskInput describes the common fields of a single node (epic, story or task)
// in a hierarchy created via eva_task_create_tree.
type TreeTaskInput struct {
	Name        string     `json:"name"`
	Text        string     `json:"text,omitempty"`
	Priority    int        `json:"priority,omitempty"`
	Deadline    string     `json:"deadline,omitempty"`
	Responsible string     `json:"responsible,omitempty"`
	Executors   StringList `json:"executors,omitempty"`
	Tags        TagList    `json:"tags,omitempty"`
	Lists       StringList `json:"lists,omitempty"`
	// Status, when set, is applied right after creation via a workflow
	// transition that preserves epic_id (e.g. "Backlog", "OPEN", "CLOSED").
	// Overrides the tree-wide default status.
	Status string `json:"status,omitempty"`
}

// TaskInputList is a []TreeTaskInput tolerant of a JSON-encoded string on input.
type TaskInputList []TreeTaskInput

func (l *TaskInputList) UnmarshalJSON(data []byte) error {
	return unmarshalFlexibleSlice(data, (*[]TreeTaskInput)(l))
}

// StoryNode is a story together with the tasks nested under it.
type StoryNode struct {
	TreeTaskInput
	Tasks TaskInputList `json:"tasks,omitempty"`
}

// StoryNodeList is a []StoryNode tolerant of a JSON-encoded string on input.
type StoryNodeList []StoryNode

func (l *StoryNodeList) UnmarshalJSON(data []byte) error {
	return unmarshalFlexibleSlice(data, (*[]StoryNode)(l))
}

// EpicNode is an epic together with the stories nested under it.
type EpicNode struct {
	TreeTaskInput
	Stories StoryNodeList `json:"stories,omitempty"`
}

// EpicNodeList is a []EpicNode tolerant of a JSON-encoded string on input.
type EpicNodeList []EpicNode

func (l *EpicNodeList) UnmarshalJSON(data []byte) error {
	return unmarshalFlexibleSlice(data, (*[]EpicNode)(l))
}

// TaskCreateTreeInput is the input for eva_task_create_tree.
//
// It creates a three-level hierarchy in a single call: one or more epics, the
// stories under each epic, and the tasks under each story. epic_id forms a chain
// up the hierarchy: a story's epic_id is its epic, and a task's epic_id is its
// story. Each task also gets parent_task set to its story.
//
// EpicLogicType/StoryLogicType/TaskLogicType accept either a logic type code
// (e.g. "task.epic:default") or a logic type ID ("CmfLogicType:uuid"). When
// empty, the well-known defaults (epic/userstory/agile-task) are used.
type TaskCreateTreeInput struct {
	ProjectID string       `json:"project_id"`
	Epics     EpicNodeList `json:"epics"`

	EpicLogicType  string `json:"epic_logic_type,omitempty"`
	StoryLogicType string `json:"story_logic_type,omitempty"`
	TaskLogicType  string `json:"task_logic_type,omitempty"`

	// Status is the default status applied to every node after creation, unless
	// the node sets its own status. Empty leaves nodes in their initial status.
	Status string `json:"status,omitempty"`

	// Workers controls parallelism for create calls (default: 3, max: 10).
	Workers int `json:"workers,omitempty"`
}

// TreeNodeResult is a compact result for a single created node. It keeps the
// payload small (the whole point of the bulk tool) by returning only identity
// and linkage fields. On failure only Name and Error are set.
type TreeNodeResult struct {
	ID     string `json:"id,omitempty"`
	Code   string `json:"code,omitempty"`
	Name   string `json:"name"`
	EpicID string `json:"epic_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

// StoryNodeResult is a created story plus the results of its tasks.
type StoryNodeResult struct {
	TreeNodeResult
	Tasks []*TreeNodeResult `json:"tasks,omitempty"`
}

// EpicNodeResult is a created epic plus the results of its stories.
type EpicNodeResult struct {
	TreeNodeResult
	Stories []*StoryNodeResult `json:"stories,omitempty"`
}

// TaskCreateTreeResult summarises the bulk-create operation.
type TaskCreateTreeResult struct {
	Epics  []*EpicNodeResult `json:"epics"`
	Total  int               `json:"total"`
	Failed int               `json:"failed,omitempty"`
}

// logicTypeIDs holds the resolved logic type IDs for each hierarchy level.
type logicTypeIDs struct {
	epic, story, task string
}

// TaskCreateTree creates one or more epics, their stories and tasks in one call,
// linking every descendant to its epic via epic_id and every task to its story.
func (t *TaskTools) TaskCreateTree(ctx context.Context, input *TaskCreateTreeInput) (any, error) {
	ltCache := make(map[string]string)
	var lt logicTypeIDs
	var err error

	lt.epic, err = t.resolveLogicTypeID(ctx, firstNonEmpty(input.EpicLogicType, evateamclient.LogicTypeCodeEpic), ltCache)
	if err != nil {
		return nil, WrapError("task_create_tree", err)
	}
	lt.story, err = t.resolveLogicTypeID(ctx, firstNonEmpty(input.StoryLogicType, evateamclient.LogicTypeCodeStory), ltCache)
	if err != nil {
		return nil, WrapError("task_create_tree", err)
	}
	lt.task, err = t.resolveLogicTypeID(ctx, firstNonEmpty(input.TaskLogicType, evateamclient.LogicTypeCodeTask), ltCache)
	if err != nil {
		return nil, WrapError("task_create_tree", err)
	}

	workers := input.Workers
	if workers <= 0 {
		workers = defaultTreeWorkers
	}
	if workers > maxTreeWorkers {
		workers = maxTreeWorkers
	}
	sem := make(chan struct{}, workers)

	epics := make([]*EpicNodeResult, len(input.Epics))
	var wg sync.WaitGroup
	for ei := range input.Epics {
		epic := &input.Epics[ei]
		epics[ei] = &EpicNodeResult{TreeNodeResult: TreeNodeResult{Name: epic.Name}}
		wg.Add(1)
		go func(ei int, epic *EpicNode) {
			defer wg.Done()
			t.createEpicSubtree(ctx, input.ProjectID, epic, lt, input.Status, sem, epics[ei])
		}(ei, epic)
	}
	wg.Wait()

	total, failed := 0, 0
	for _, e := range epics {
		total++
		if e.Error != "" {
			failed++
		}
		for _, s := range e.Stories {
			total++
			if s.Error != "" {
				failed++
			}
			for _, task := range s.Tasks {
				total++
				if task.Error != "" {
					failed++
				}
			}
		}
	}

	return &TaskCreateTreeResult{
		Epics:  epics,
		Total:  total,
		Failed: failed,
	}, nil
}

// createEpicSubtree creates an epic and, on success, its stories (and their
// tasks) in parallel, writing all results into out.
func (t *TaskTools) createEpicSubtree(
	ctx context.Context,
	projectID string,
	epic *EpicNode,
	lt logicTypeIDs,
	defaultStatus string,
	sem chan struct{},
	out *EpicNodeResult,
) {
	epicTask, err := t.createTreeNode(ctx, treeParams(&epic.TreeTaskInput, projectID, lt.epic, "", ""),
		firstNonEmpty(epic.Status, defaultStatus), sem)
	if epicTask == nil {
		out.Error = err.Error()
		return
	}
	out.TreeNodeResult = TreeNodeResult{ID: epicTask.ID, Code: epicTask.Code, Name: epicTask.Name}
	if err != nil {
		out.Error = err.Error() // created, but status transition failed
	}

	out.Stories = make([]*StoryNodeResult, len(epic.Stories))
	var wg sync.WaitGroup
	for si := range epic.Stories {
		story := &epic.Stories[si]
		out.Stories[si] = &StoryNodeResult{TreeNodeResult: TreeNodeResult{Name: story.Name}}
		wg.Add(1)
		go func(si int, story *StoryNode) {
			defer wg.Done()
			t.createStorySubtree(ctx, projectID, story, lt, defaultStatus, epicTask.ID, sem, out.Stories[si])
		}(si, story)
	}
	wg.Wait()
}

// createStorySubtree creates a story and, on success, its tasks in parallel,
// writing all results into out.
func (t *TaskTools) createStorySubtree(
	ctx context.Context,
	projectID string,
	story *StoryNode,
	lt logicTypeIDs,
	defaultStatus string,
	epicID string,
	sem chan struct{},
	out *StoryNodeResult,
) {
	storyTask, err := t.createTreeNode(ctx, treeParams(&story.TreeTaskInput, projectID, lt.story, epicID, ""),
		firstNonEmpty(story.Status, defaultStatus), sem)
	if storyTask == nil {
		out.Error = err.Error()
		return
	}
	out.TreeNodeResult = TreeNodeResult{ID: storyTask.ID, Code: storyTask.Code, Name: storyTask.Name, EpicID: epicID}
	if err != nil {
		out.Error = err.Error() // created, but status transition failed
	}

	out.Tasks = make([]*TreeNodeResult, len(story.Tasks))
	var wg sync.WaitGroup
	for ti := range story.Tasks {
		task := &story.Tasks[ti]
		out.Tasks[ti] = &TreeNodeResult{Name: task.Name}
		wg.Add(1)
		go func(ti int, task *TreeTaskInput) {
			defer wg.Done()

			// epic_id forms a chain: a task's parent agile item is its story,
			// not the root epic (the story itself links up to the epic).
			created, taskErr := t.createTreeNode(ctx, treeParams(task, projectID, lt.task, storyTask.ID, storyTask.ID),
				firstNonEmpty(task.Status, defaultStatus), sem)
			if created == nil {
				out.Tasks[ti].Error = taskErr.Error()
				return
			}
			out.Tasks[ti] = &TreeNodeResult{ID: created.ID, Code: created.Code, Name: created.Name, EpicID: storyTask.ID}
			if taskErr != nil {
				out.Tasks[ti].Error = taskErr.Error() // created, but status transition failed
			}
		}(ti, task)
	}
	wg.Wait()
}

// createTreeNode creates one task and, when status is non-empty, applies it via
// TaskUpdateStatus (which preserves epic_id across the workflow transition).
// Concurrency is bounded by sem. On create failure it returns (nil, err); on
// status failure it returns the created task together with the status error.
func (t *TaskTools) createTreeNode(
	ctx context.Context,
	params *evateamclient.TaskCreateParams,
	status string,
	sem chan struct{},
) (*models.Task, error) {
	sem <- struct{}{}
	task, err := t.client.TaskCreate(ctx, params)
	<-sem
	if err != nil {
		return nil, err
	}
	if status == "" {
		return task, nil
	}

	sem <- struct{}{}
	updated, statusErr := t.client.TaskUpdateStatus(ctx, task.ID, status)
	<-sem
	if statusErr != nil {
		return task, statusErr
	}
	return updated, nil
}

// resolveLogicTypeID maps a logic type code to its ID, caching results. A value
// that already looks like an ID ("CmfLogicType:uuid") is returned unchanged.
func (t *TaskTools) resolveLogicTypeID(ctx context.Context, codeOrID string, cache map[string]string) (string, error) {
	if strings.HasPrefix(codeOrID, "CmfLogicType:") {
		return codeOrID, nil
	}
	if id, ok := cache[codeOrID]; ok {
		return id, nil
	}
	lt, err := t.client.LogicTypeByCode(ctx, codeOrID)
	if err != nil {
		return "", err
	}
	cache[codeOrID] = lt.ID
	return lt.ID, nil
}

// treeParams builds TaskCreateParams for a node, wiring its epic and parent links.
func treeParams(node *TreeTaskInput, projectID, logicTypeID, epicID, parentTask string) *evateamclient.TaskCreateParams {
	return &evateamclient.TaskCreateParams{
		Name:        node.Name,
		ProjectID:   projectID,
		Text:        node.Text,
		Priority:    node.Priority,
		Deadline:    node.Deadline,
		Responsible: node.Responsible,
		Executors:   []string(node.Executors),
		Tags:        []string(node.Tags),
		Lists:       []string(node.Lists),
		Epic:        epicID,
		ParentTask:  parentTask,
		LogicTypeID: logicTypeID,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// taskCreateTreeInputSchema builds the input JSON schema and relaxes every
// array-typed node to also accept a JSON-encoded string, mirroring the flexible
// unmarshalling of the *List types. Some MCP clients send (nested) array
// arguments as strings; without this the SDK rejects them before the handler
// runs. Panics on failure since the type is fixed and reflection is deterministic.
func taskCreateTreeInputSchema() *jsonschema.Schema {
	s, err := jsonschema.ForType(reflect.TypeFor[TaskCreateTreeInput](), &jsonschema.ForOptions{})
	if err != nil {
		panic("build eva_task_create_tree input schema: " + err.Error())
	}
	relaxArraySchema(s)
	return s
}

// relaxArraySchema recursively adds "string" as an accepted type to every
// array-typed schema node.
func relaxArraySchema(s *jsonschema.Schema) {
	if s == nil {
		return
	}
	switch {
	case s.Type == "array":
		s.Type = ""
		s.Types = []string{"array", "string"}
	case slices.Contains(s.Types, "array") && !slices.Contains(s.Types, "string"):
		s.Types = append(s.Types, "string")
	}
	relaxArraySchema(s.Items)
	relaxArraySchema(s.AdditionalProperties)
	for _, p := range s.Properties {
		relaxArraySchema(p)
	}
}
