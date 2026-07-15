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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raoptimus/evateamclient.go/models"
)

// This file contains executable reproductions of the bugs collected while
// operating the evateam MCP server (report categories B, C, D, E). They assert
// the DESIRED invariants; the ones covering server-side defects (B/C/D) are
// expected to FAIL against the current server and serve as a runnable source of
// truth for the server team. E1 is client-fixable and goes green after the fix.
//
// Gated by EVA_API_URL / EVA_API_TOKEN, skipped by `make test` (name prefix).

// epicStoryTask creates an epic → story → task chain and registers cleanup.
// The task is placed under the story (epic_id chains up to the story).
func epicStoryTask(t *testing.T, c *Client, projectID string) (epic, story, task *models.Task) {
	t.Helper()
	ctx := context.Background()

	epicLT, err := c.LogicTypeByCode(ctx, LogicTypeCodeEpic)
	require.NoError(t, err)
	storyLT, err := c.LogicTypeByCode(ctx, LogicTypeCodeStory)
	require.NoError(t, err)
	taskLT, err := c.LogicTypeByCode(ctx, LogicTypeCodeTask)
	require.NoError(t, err)

	suffix := time.Now().UnixNano()

	epic, err = c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] bug-epic %d", suffix),
		ProjectID:   projectID,
		LogicTypeID: epicLT.ID,
		Priority:    1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, epic.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), epic.ID) })

	story, err = c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] bug-story %d", suffix),
		ProjectID:   projectID,
		Epic:        epic.ID,
		LogicTypeID: storyLT.ID,
		Priority:    2,
	})
	require.NoError(t, err)
	require.NotEmpty(t, story.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), story.ID) })

	task, err = c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] bug-task %d", suffix),
		ProjectID:   projectID,
		Epic:        story.ID,
		ParentTask:  story.ID,
		LogicTypeID: taskLT.ID,
		Priority:    3,
	})
	require.NoError(t, err)
	require.NotEmpty(t, task.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), task.ID) })

	return epic, story, task
}

// TestIntegration_Bug_B1_TextUpdateResetsEpic reproduces report B1: updating
// only `text` (without re-sending epic_id) server-side detaches the task from
// its parent story. Desired invariant: epic_id survives a text-only update.
func TestIntegration_Bug_B1_TextUpdateResetsEpic(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	_, story, task := epicStoryTask(t, c, projectID)

	before, _, err := c.EpicByID(ctx, task.ID, []string{TaskFieldID, TaskFieldEpicID})
	require.NoError(t, err)
	require.Equal(t, story.ID, before.EpicID, "precondition: task hangs under the story")

	_, err = c.TaskUpdate(ctx, task.ID, map[string]any{"text": "updated body only"})
	require.NoError(t, err)

	after, _, err := c.EpicByID(ctx, task.ID, []string{TaskFieldID, TaskFieldEpicID})
	require.NoError(t, err)
	assert.Equal(t, story.ID, after.EpicID,
		"KNOWN SERVER BUG B1: text-only update must not detach the task from its story")
}

// TestIntegration_Bug_B2_TextUpdateResetsResponsibleAndStatus reproduces report
// B2: a name/text update also resets responsible_id and status. Desired
// invariant: sibling fields are untouched by a partial update.
func TestIntegration_Bug_B2_TextUpdateResetsResponsibleAndStatus(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	_, _, task := epicStoryTask(t, c, projectID)

	readFields := []string{TaskFieldID, TaskFieldResponsibleID, TaskFieldCacheStatusType, TaskFieldStatusID}
	before, _, err := c.EpicByID(ctx, task.ID, readFields)
	require.NoError(t, err)

	_, err = c.TaskUpdate(ctx, task.ID, map[string]any{
		"name": "[TEST] renamed",
		"text": "new body",
	})
	require.NoError(t, err)

	after, _, err := c.EpicByID(ctx, task.ID, readFields)
	require.NoError(t, err)
	assert.Equal(t, before.ResponsibleID, after.ResponsibleID,
		"KNOWN SERVER BUG B2: name/text update must not change responsible_id")
	assert.Equal(t, before.CacheStatusType, after.CacheStatusType,
		"KNOWN SERVER BUG B2: name/text update must not change status")
	assert.Equal(t, before.StatusID, after.StatusID,
		"KNOWN SERVER BUG B2: name/text update must not change status_id")
}

// TestIntegration_Bug_B3_FieldOnlyUpdateChangesStatus reproduces report B3: a
// field-only update (epic_id) unexpectedly changes the status. Desired
// invariant: status is untouched when it is not part of the update.
func TestIntegration_Bug_B3_FieldOnlyUpdateChangesStatus(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	epic, _, task := epicStoryTask(t, c, projectID)

	readFields := []string{TaskFieldID, TaskFieldCacheStatusType, TaskFieldStatusID}
	before, _, err := c.EpicByID(ctx, task.ID, readFields)
	require.NoError(t, err)

	// Re-parent the task directly under the epic (a field-only update).
	_, err = c.TaskUpdate(ctx, task.ID, map[string]any{"epic": epic.ID})
	require.NoError(t, err)

	after, _, err := c.EpicByID(ctx, task.ID, readFields)
	require.NoError(t, err)
	assert.Equal(t, before.CacheStatusType, after.CacheStatusType,
		"KNOWN SERVER BUG B3: field-only update must not change status")
	assert.Equal(t, before.StatusID, after.StatusID,
		"KNOWN SERVER BUG B3: field-only update must not change status_id")
}

// TestIntegration_Bug_C1_EpicIDDenormalizedToRootEpic reproduces report C1: for
// a task under a Story, epic_id is denormalized to the ROOT epic instead of the
// immediate parent story. Desired invariant: epic_id is the immediate parent.
func TestIntegration_Bug_C1_EpicIDDenormalizedToRootEpic(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	_, story, task := epicStoryTask(t, c, projectID)

	got, _, err := c.EpicByID(ctx, task.ID, []string{TaskFieldID, TaskFieldEpicID})
	require.NoError(t, err)
	assert.Equal(t, story.ID, got.EpicID,
		"KNOWN SERVER BUG C1: epic_id must point at the immediate parent story, not the root epic")
}

// TestIntegration_Bug_D2_StatusIDReadOnly reproduces report D2: setting a
// concrete substatus via updates.status_id is rejected server-side with
// "CmfTask.update returned empty id".
func TestIntegration_Bug_D2_StatusIDReadOnly(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	_, _, task := epicStoryTask(t, c, projectID)

	cur, _, err := c.EpicByID(ctx, task.ID, []string{TaskFieldID, TaskFieldStatusID})
	require.NoError(t, err)
	require.NotEmpty(t, cur.StatusID)

	_, err = c.TaskUpdate(ctx, task.ID, map[string]any{"status_id": cur.StatusID})
	// Characterizes the current behavior: status_id is readonly and the update
	// returns an empty id, surfaced as an error by the client.
	assert.Error(t, err, "KNOWN SERVER BUG D2: status_id is readonly and update returns empty id")
}

// TestIntegration_Bug_E1_GetMissingCodeReturnsEmpty reproduces report E1 at the
// client-library level: c.Task for a non-existent code returns an empty object
// instead of an error (the caller must check .code != ""). The MCP handler fix
// lives one layer up — see TestIntegration_TaskGet_MissingCodeReturnsError in
// the tools package, where eva_task_get now surfaces ErrNotFound.
func TestIntegration_Bug_E1_GetMissingCodeReturnsEmpty(t *testing.T) {
	c := newIntegrationClient(t)
	_ = getIntegrationProjectID(t, c)
	ctx := context.Background()

	task, _, err := c.Task(ctx, "NOPE-00000000", []string{TaskFieldID, TaskFieldCode, TaskFieldName})
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Empty(t, task.Code, "non-existent code yields an empty task object (caller must check .code != \"\")")
}
