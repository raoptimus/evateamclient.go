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

func TestIntegration_TaskUpdate_EpicAndStory(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	epicLT, err := c.LogicTypeByCode(ctx, LogicTypeCodeEpic)
	require.NoError(t, err)
	storyLT, err := c.LogicTypeByCode(ctx, LogicTypeCodeStory)
	require.NoError(t, err)

	suffix := time.Now().UnixNano()

	epic, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] update-epic original %d", suffix),
		ProjectID:   projectID,
		Text:        "original epic text",
		LogicTypeID: epicLT.ID,
		Priority:    1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, epic.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), epic.ID) })

	story, err := c.TaskCreate(ctx, &TaskCreateParams{
		Name:        fmt.Sprintf("[TEST] update-story original %d", suffix),
		ProjectID:   projectID,
		Text:        "original story text",
		Epic:        epic.ID,
		LogicTypeID: storyLT.ID,
		Priority:    2,
	})
	require.NoError(t, err)
	require.NotEmpty(t, story.ID)
	t.Cleanup(func() { _ = c.TaskDelete(context.Background(), story.ID) })

	t.Run("update epic name+text+priority", func(t *testing.T) {
		newName := fmt.Sprintf("[TEST] update-epic UPDATED %d", suffix)
		updated, err := c.TaskUpdate(ctx, epic.ID, map[string]any{
			"name":     newName,
			"text":     "updated epic text",
			"priority": 5,
		})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, epic.ID, updated.ID, "update should return same task ID")
		assert.Equal(t, newName, updated.Name)
		assert.Equal(t, "updated epic text", updated.Text)
		assert.Equal(t, 5, updated.Priority)

		// Re-fetch independently to confirm persistence.
		fetched, _, err := c.EpicByID(ctx, epic.ID, []string{
			TaskFieldID, TaskFieldName, TaskFieldText, TaskFieldPriority,
		})
		require.NoError(t, err)
		assert.Equal(t, newName, fetched.Name)
		assert.Equal(t, "updated epic text", fetched.Text)
		assert.Equal(t, 5, fetched.Priority)
	})

	t.Run("update story re-parents to no epic then back", func(t *testing.T) {
		// Drop epic link.
		_, err := c.TaskUpdate(ctx, story.ID, map[string]any{
			"epic": nil,
		})
		require.NoError(t, err)

		fetched, _, err := c.EpicByID(ctx, story.ID, []string{
			TaskFieldID, TaskFieldEpicID,
		})
		require.NoError(t, err)
		assert.Empty(t, fetched.EpicID, "epic link should be cleared")

		// Re-link to epic.
		_, err = c.TaskUpdate(ctx, story.ID, map[string]any{
			"epic": epic.ID,
		})
		require.NoError(t, err)

		fetched, _, err = c.EpicByID(ctx, story.ID, []string{
			TaskFieldID, TaskFieldEpicID,
		})
		require.NoError(t, err)
		assert.Equal(t, epic.ID, fetched.EpicID, "epic link should be restored")
	})
}
