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
)

func TestIntegration_TaskCreate_EpicStoryTaskBug(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	logicTypes := map[string]string{}
	for _, code := range []string{
		LogicTypeCodeEpic,
		LogicTypeCodeStory,
		LogicTypeCodeTask,
		LogicTypeCodeBug,
	} {
		lt, err := c.LogicTypeByCode(ctx, code)
		require.NoError(t, err, "logic type %q must exist on server", code)
		require.NotEmpty(t, lt.ID)
		logicTypes[code] = lt.ID
	}

	verifyFields := []string{
		TaskFieldID,
		TaskFieldCode,
		TaskFieldName,
		TaskFieldEpicID,
		TaskFieldParentTaskID,
		TaskFieldLogicType,
	}

	suffix := time.Now().UnixNano()

	epic, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] integration epic %d", suffix),
		ProjectID:   projectID,
		Text:        "created by TestIntegration_TaskCreate_EpicStoryTaskBug",
		LogicTypeID: logicTypes[LogicTypeCodeEpic],
	})
	require.NoError(t, err)
	require.NotNil(t, epic)
	require.NotEmpty(t, epic.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), epic.ID) })

	story, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] integration story %d", suffix),
		ProjectID:   projectID,
		Text:        "child of integration epic",
		Epic:        epic.ID,
		LogicTypeID: logicTypes[LogicTypeCodeStory],
	})
	require.NoError(t, err)
	require.NotNil(t, story)
	require.NotEmpty(t, story.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), story.ID) })

	task, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] integration task %d", suffix),
		ProjectID:   projectID,
		Text:        "child task of integration story",
		ParentTask:  story.ID,
		LogicTypeID: logicTypes[LogicTypeCodeTask],
	})
	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotEmpty(t, task.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), task.ID) })

	bug, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] integration bug %d", suffix),
		ProjectID:   projectID,
		Text:        "child bug of integration story",
		ParentTask:  story.ID,
		LogicTypeID: logicTypes[LogicTypeCodeBug],
	})
	require.NoError(t, err)
	require.NotNil(t, bug)
	require.NotEmpty(t, bug.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), bug.ID) })

	// Verify 1: story is linked to epic and has Story logic type.
	fetchedStory, _, err := c.EpicByID(ctx, story.ID, verifyFields)
	require.NoError(t, err)
	require.NotNil(t, fetchedStory)
	assert.Equal(t, epic.ID, fetchedStory.EpicID, "story should be linked to epic")
	require.NotNil(t, fetchedStory.LogicType)
	assert.Equal(t, LogicTypeCodeStory, fetchedStory.LogicType.Code)

	// Verify 2: story appears in EpicTasks(epic.ID).
	children, _, err := c.EpicTasks(ctx, epic.ID, DefaultEpicListFields)
	require.NoError(t, err)
	foundStory := false
	for _, ch := range children {
		if ch.ID == story.ID {
			foundStory = true
			break
		}
	}
	assert.True(t, foundStory, "story %s should appear in EpicTasks(%s)", story.ID, epic.ID)

	// Verify 3: task and bug are linked to story via ParentTaskID and have correct logic types.
	for _, child := range []struct {
		name     string
		id       string
		typeCode string
	}{
		{"task", task.ID, LogicTypeCodeTask},
		{"bug", bug.ID, LogicTypeCodeBug},
	} {
		t.Run("child_"+child.name, func(t *testing.T) {
			fetched, _, err := c.EpicByID(ctx, child.id, verifyFields)
			require.NoError(t, err)
			require.NotNil(t, fetched)
			assert.Equal(t, story.ID, fetched.ParentTaskID, "%s should be linked to story", child.name)
			require.NotNil(t, fetched.LogicType)
			assert.Equal(t, child.typeCode, fetched.LogicType.Code)
		})
	}
}
