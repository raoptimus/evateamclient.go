/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TaskLinksListQuery_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityRelation).
		Limit(5)

	links, meta, err := c.TaskLinksListQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, links, "should have task links")

	for _, link := range links {
		assert.True(t, strings.HasPrefix(link.ID, "CmfRelationOption:"),
			"ID should start with CmfRelationOption:, got %s", link.ID)
	}
}

func TestIntegration_TaskLinksListQuery_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultTaskLinkListFields...).
		From(EntityRelation).
		Limit(5)

	links, meta, err := c.TaskLinksListQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, links)

	for _, link := range links {
		assert.True(t, strings.HasPrefix(link.ID, "CmfRelationOption:"))
		assert.NotEmpty(t, link.Code)
	}
}

func TestIntegration_TaskLink_ByID(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	// Get a task link ID from the list first
	qb := NewQueryBuilder().
		Select(TaskLinkFieldID).
		From(EntityRelation).
		Limit(1)

	links, _, err := c.TaskLinksListQuery(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, links, "need at least one task link")

	linkID := links[0].ID
	require.NotEmpty(t, linkID)

	// Fetch single task link by ID with default fields
	link, meta, err := c.TaskLink(ctx, linkID, DefaultTaskLinkFields)
	require.NoError(t, err)
	require.NotNil(t, link)
	require.NotNil(t, meta)

	assert.Equal(t, linkID, link.ID)
	assert.NotEmpty(t, link.Code)
	assert.False(t, link.CmfCreatedAt.IsZero(), "CmfCreatedAt should not be zero")
}

func TestIntegration_TaskLinkCount(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(TaskLinkFieldID).
		From(EntityRelation)

	count, err := c.TaskLinkCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "should have task links")
}

func TestIntegration_TaskLinksOutgoing(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a task ID from the project
	taskQB := NewQueryBuilder().
		Select(TaskFieldID).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, taskQB)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "need at least one task in project")

	taskID := tasks[0].ID
	require.NotEmpty(t, taskID)

	// Get outgoing links for the task (may be empty)
	links, meta, err := c.TaskLinksOutgoing(ctx, taskID, DefaultTaskLinkListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(links) == 0 {
		t.Skip("no outgoing links found for task, skipping detailed assertions")
	}

	for _, link := range links {
		assert.True(t, strings.HasPrefix(link.ID, "CmfRelationOption:"))
	}
}

// TestIntegration_TaskLinkCreate_VisibleViaBothReadPaths is the red-green loop
// for the original bug: creating a link must actually create it, and the new
// link must be readable back both via CmfRelationOption.list (TaskLinksOutgoing/
// TaskLinksIncoming) and via the documented KB-000325 path (CmfTask.get with
// fields ["out_tasks.**","in_tasks.**"]). Deletes the link and both tasks.
func TestIntegration_TaskLinkCreate_VisibleViaBothReadPaths(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	lt, err := c.LogicTypeByCode(ctx, LogicTypeCodeTask)
	require.NoError(t, err)

	suffix := time.Now().UnixNano()

	taskOut, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] tasklink out %d", suffix),
		ProjectID:   projectID,
		LogicTypeID: lt.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, taskOut.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), taskOut.ID) })

	taskIn, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] tasklink in %d", suffix),
		ProjectID:   projectID,
		LogicTypeID: lt.ID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, taskIn.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), taskIn.ID) })

	link, err := c.TaskLinkCreate(ctx, taskOut.ID, taskIn.ID, RelationTypeLink)
	require.NoError(t, err)
	require.NotNil(t, link)
	require.NotEmpty(t, link.ID, "TaskLinkCreate must return a created link, not a silent no-op")

	linkDeleted := false
	t.Cleanup(func() {
		if !linkDeleted {
			_ = c.TaskLinkDelete(context.Background(), link.ID)
		}
	})

	t.Run("read_via_relation_list", func(t *testing.T) {
		outgoing, _, err := c.TaskLinksOutgoing(ctx, taskOut.ID, DefaultTaskLinkListFields)
		require.NoError(t, err)
		assert.True(t, containsTaskLinkID(outgoing, link.ID),
			"created link should appear in TaskLinksOutgoing(%s)", taskOut.ID)

		incoming, _, err := c.TaskLinksIncoming(ctx, taskIn.ID, DefaultTaskLinkListFields)
		require.NoError(t, err)
		assert.True(t, containsTaskLinkID(incoming, link.ID),
			"created link should appear in TaskLinksIncoming(%s)", taskIn.ID)
	})

	t.Run("read_via_task_out_in_tasks_KB-000325", func(t *testing.T) {
		qb := NewQueryBuilder().
			From(EntityTask).
			Where(sq.Eq{TaskFieldID: taskOut.ID}).
			Limit(1)
		kwargs, err := qb.ToKwargs()
		require.NoError(t, err)
		kwargs["fields"] = []string{"out_tasks.**", "in_tasks.**"}

		reqBody := &RPCRequest{
			JSONRPC: "2.2",
			Method:  "CmfTask.get",
			CallID:  newCallID(),
			Kwargs:  kwargs,
		}

		var resp struct {
			Result struct {
				OutTasks []models.TaskLink `json:"out_tasks"`
				InTasks  []models.TaskLink `json:"in_tasks"`
			} `json:"result"`
		}
		require.NoError(t, c.doRequest(ctx, reqBody, &resp))
		// KB-000325 does not guarantee which side the server files the link
		// under; assert presence in either collection to avoid a false red.
		found := containsTaskLinkID(resp.Result.OutTasks, link.ID) ||
			containsTaskLinkID(resp.Result.InTasks, link.ID)
		assert.True(t, found,
			"KB-000325: created link should appear in CmfTask.get(%s).out_tasks or .in_tasks", taskOut.ID)
	})

	require.NoError(t, c.TaskLinkDelete(ctx, link.ID))
	linkDeleted = true

	_, _, err = c.TaskLink(ctx, link.ID, nil)
	assert.Error(t, err, "deleted link should no longer be readable")
}

func containsTaskLinkID(links []models.TaskLink, id string) bool {
	for _, l := range links {
		if l.ID == id {
			return true
		}
	}
	return false
}

func TestIntegration_TaskLinks_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	kwargs := map[string]any{
		"slice": []int{0, 5},
	}

	links, meta, err := c.TaskLinksList(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, links)

	for _, link := range links {
		assert.True(t, strings.HasPrefix(link.ID, "CmfRelationOption:"))
	}
}
