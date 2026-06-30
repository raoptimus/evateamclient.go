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
	"sync"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/models"
)

// ChildTaskInput describes a single child task to create under the parent.
type ChildTaskInput struct {
	Name        string   `json:"name"`
	Text        string   `json:"text,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	Responsible string   `json:"responsible,omitempty"`
	Executors   []string `json:"executors,omitempty"`
	Tags        TagList  `json:"tags,omitempty"`
	Lists       []string `json:"lists,omitempty"`
	LogicTypeID string   `json:"logic_type_id,omitempty"`
}

// TaskCreateWithSubtasksInput is the input for eva_task_create_with_subtasks.
type TaskCreateWithSubtasksInput struct {
	Name        string           `json:"name"`
	ProjectID   string           `json:"project_id"`
	Text        string           `json:"text,omitempty"`
	Priority    int              `json:"priority,omitempty"`
	Deadline    string           `json:"deadline,omitempty"`
	Responsible string           `json:"responsible,omitempty"`
	Executors   []string         `json:"executors,omitempty"`
	Tags        TagList          `json:"tags,omitempty"`
	Lists       []string         `json:"lists,omitempty"`
	Epic        string           `json:"epic,omitempty"`
	LogicTypeID string           `json:"logic_type_id,omitempty"`
	Children    []ChildTaskInput `json:"children"`
	// Workers controls parallelism for child creation (default: 3, max: 10).
	Workers int `json:"workers,omitempty"`
}

// ChildTaskResult embeds the created Task and carries an error message if creation failed.
// On failure, only Name (from input) and Error are set.
type ChildTaskResult struct {
	models.Task
	Error string `json:"error,omitempty"`
}

// TaskCreateWithSubtasksResult summarises the bulk-create operation.
type TaskCreateWithSubtasksResult struct {
	Parent   *models.Task       `json:"parent"`
	Children []*ChildTaskResult `json:"children"`
	Total    int                `json:"total"`
	Failed   int                `json:"failed,omitempty"`
}

// TaskCreateWithSubtasks creates a parent task then all child tasks in parallel.
func (t *TaskTools) TaskCreateWithSubtasks(ctx context.Context, input *TaskCreateWithSubtasksInput) (any, error) {
	parent, err := t.client.TaskCreate(ctx, &evateamclient.TaskCreateParams{
		Name:        input.Name,
		ProjectID:   input.ProjectID,
		Text:        input.Text,
		Priority:    input.Priority,
		Deadline:    input.Deadline,
		Responsible: input.Responsible,
		Executors:   input.Executors,
		Tags:        []string(input.Tags),
		Lists:       input.Lists,
		Epic:        input.Epic,
		LogicTypeID: input.LogicTypeID,
	})
	if err != nil {
		return nil, WrapError("task_create_with_subtasks", err)
	}

	workers := input.Workers
	if workers <= 0 {
		workers = defaultTreeWorkers
	}
	if workers > maxTreeWorkers {
		workers = maxTreeWorkers
	}

	results := make([]*ChildTaskResult, len(input.Children))
	for i := range input.Children {
		results[i] = &ChildTaskResult{Task: models.Task{TaskBrowse: models.TaskBrowse{Name: input.Children[i].Name}}}
	}

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i := range input.Children {
		wg.Add(1)
		go func(i int, child ChildTaskInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			task, createErr := t.client.TaskCreate(ctx, &evateamclient.TaskCreateParams{
				Name:        child.Name,
				ProjectID:   input.ProjectID,
				Text:        child.Text,
				Priority:    child.Priority,
				Deadline:    child.Deadline,
				Responsible: child.Responsible,
				Executors:   child.Executors,
				Tags:        []string(child.Tags),
				Lists:       child.Lists,
				ParentTask:  parent.ID,
				LogicTypeID: child.LogicTypeID,
			})
			if createErr != nil {
				results[i].Error = createErr.Error()
			} else {
				results[i].Task = *task
			}
		}(i, input.Children[i])
	}

	wg.Wait()

	failed := 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		}
	}

	return &TaskCreateWithSubtasksResult{
		Parent:   parent,
		Children: results,
		Total:    len(input.Children),
		Failed:   failed,
	}, nil
}
