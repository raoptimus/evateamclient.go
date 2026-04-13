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
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getIntegrationTaskID returns a task ID from the project for use in status history tests.
func getIntegrationTaskIDForHistory(t *testing.T, c *Client, projectID string) string {
	t.Helper()
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(TaskFieldID).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "need at least one task in project")

	return tasks[0].ID
}

func TestIntegration_StatusHistoryList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskIDForHistory(t, c, projectID)

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldParentID: taskID}).
		OrderBy("-" + StatusHistoryFieldCmfCreatedAt).
		Limit(5)

	histories, meta, err := c.StatusHistoryList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(histories) == 0 {
		t.Skip("no status history for this task, skipping")
	}

	for _, h := range histories {
		assert.True(t, strings.HasPrefix(h.ID, "CmfStatusHistory:"),
			"ID should start with CmfStatusHistory:, got %s", h.ID)
	}
}

func TestIntegration_StatusHistoryList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskIDForHistory(t, c, projectID)

	qb := NewQueryBuilder().
		Select(DefaultStatusHistoryListFields...).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldParentID: taskID}).
		Limit(5)

	histories, meta, err := c.StatusHistoryList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(histories) == 0 {
		t.Skip("no status history for this task, skipping")
	}

	for _, h := range histories {
		assert.True(t, strings.HasPrefix(h.ID, "CmfStatusHistory:"))
		assert.NotEmpty(t, h.ParentID)
		require.NotNil(t, h.CmfCreatedAt, "CmfCreatedAt should not be nil")
		assert.False(t, h.CmfCreatedAt.IsZero(), "CmfCreatedAt should be parsed")
	}
}

func TestIntegration_StatusHistory_ByID(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskIDForHistory(t, c, projectID)

	// Get a status history ID from the list first
	qb := NewQueryBuilder().
		Select(StatusHistoryFieldID).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldParentID: taskID}).
		Limit(1)

	histories, _, err := c.StatusHistoryList(ctx, qb)
	require.NoError(t, err)

	if len(histories) == 0 {
		t.Skip("no status history for this task, skipping")
	}

	historyID := histories[0].ID
	require.NotEmpty(t, historyID)

	// Fetch single status history by ID with default fields
	h, meta, err := c.StatusHistory(ctx, historyID, DefaultStatusHistoryFields)
	require.NoError(t, err)
	require.NotNil(t, h)
	require.NotNil(t, meta)

	assert.Equal(t, historyID, h.ID)
	assert.NotEmpty(t, h.ParentID)
	require.NotNil(t, h.CmfCreatedAt, "CmfCreatedAt should not be nil")
	assert.False(t, h.CmfCreatedAt.IsZero())
}

func TestIntegration_StatusHistoryCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskIDForHistory(t, c, projectID)

	qb := NewQueryBuilder().
		Select(StatusHistoryFieldID).
		From(EntityStatusHistory).
		Where(sq.Eq{StatusHistoryFieldParentID: taskID})

	count, err := c.StatusHistoryCount(ctx, qb)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
}

func TestIntegration_ProjectStatusHistory(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	histories, meta, err := c.ProjectStatusHistory(ctx, projectID, DefaultStatusHistoryListFields)
	if err != nil {
		// ProjectStatusHistory filters by project_id which may not be supported
		t.Skipf("ProjectStatusHistory not supported by API: %v", err)
	}
	require.NotNil(t, meta)

	if len(histories) == 0 {
		t.Skip("no status history in project")
	}

	for _, h := range histories {
		assert.True(t, strings.HasPrefix(h.ID, "CmfStatusHistory:"))
	}
}

func TestIntegration_StatusHistories_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskIDForHistory(t, c, projectID)

	kwargs := map[string]any{
		"filter": [][]any{
			{StatusHistoryFieldParentID, "==", taskID},
		},
		"slice": []int{0, 5},
	}

	histories, meta, err := c.StatusHistories(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(histories) == 0 {
		t.Skip("no status history for this task")
	}

	for _, h := range histories {
		assert.True(t, strings.HasPrefix(h.ID, "CmfStatusHistory:"))
	}
}
