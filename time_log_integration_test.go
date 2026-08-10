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

func TestIntegration_TimeLogsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldProjectID: projectID}).
		OrderBy("-" + TimeLogFieldCmfCreatedAt).
		Limit(5)

	logs, meta, err := c.TimeLogsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, logs, "project %s should have time logs", integrationProjectCode)

	for _, log := range logs {
		assert.True(t, strings.HasPrefix(log.ID, "CmfTimeTrackerHistory:"),
			"ID should start with CmfTimeTrackerHistory:, got %s", log.ID)
		assert.Greater(t, log.TimeSpent, 0, "TimeSpent should be positive")
	}
}

func TestIntegration_TimeLogsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultTimeLogListFields...).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldProjectID: projectID}).
		Limit(5)

	logs, meta, err := c.TimeLogsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, logs)

	for _, log := range logs {
		assert.True(t, strings.HasPrefix(log.ID, "CmfTimeTrackerHistory:"))
		assert.NotEmpty(t, log.Code)
		assert.Greater(t, log.TimeSpent, 0)
		require.NotNil(t, log.CreatedAt, "CreatedAt should not be nil")
		assert.False(t, log.CreatedAt.IsZero(), "CreatedAt should be parsed")
	}
}

func TestIntegration_TimeLog_ByID(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a time log ID from the list first
	qb := NewQueryBuilder().
		Select(TimeLogFieldID).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldProjectID: projectID}).
		Limit(1)

	logs, _, err := c.TimeLogsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, logs, "need at least one time log")

	timeLogID := logs[0].ID
	require.NotEmpty(t, timeLogID)

	// Fetch single time log by ID with default fields
	log, meta, err := c.TimeLog(ctx, timeLogID, DefaultTimeLogFields)
	require.NoError(t, err)
	require.NotNil(t, log)
	require.NotNil(t, meta)

	assert.Equal(t, timeLogID, log.ID)
	assert.Greater(t, log.TimeSpent, 0)
	assert.NotEmpty(t, log.ParentID)
}

func TestIntegration_TimeLogCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(TimeLogFieldID).
		From(EntityTimeLog).
		Where(sq.Eq{TimeLogFieldProjectID: projectID})

	count, err := c.TimeLogCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "project %s should have time logs", integrationProjectCode)
}

func TestIntegration_ProjectTimeLogs(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	logs, meta, err := c.ProjectTimeLogs(ctx, projectID, DefaultTimeLogListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, logs)

	for _, log := range logs {
		assert.True(t, strings.HasPrefix(log.ID, "CmfTimeTrackerHistory:"))
		assert.Greater(t, log.TimeSpent, 0)
	}
}

// TestIntegration_TimeLogCreate_UpdateDelete_FullCycle is the red-green
// regression test for SPEC-02: TimeLogCreate/TimeLogUpdate/TimeLogDelete
// must actually write, not silently no-op (id-in-kwargs bug) or fail to
// parse a bare-string result (two-phase parsing bug).
func TestIntegration_TimeLogCreate_UpdateDelete_FullCycle(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)

	log, err := c.TimeLogCreate(ctx, TimeLogCreateParams{
		ParentID:  taskID,
		TimeSpent: 30,
	})
	require.NoError(t, err)
	require.NotNil(t, log)
	require.NotEmpty(t, log.ID, "TimeLogCreate must return a created time log, not a silent no-op")

	deleted := false
	t.Cleanup(func() {
		if !deleted {
			_ = c.TimeLogDelete(context.Background(), log.ID)
		}
	})

	fetched, _, err := c.TimeLog(ctx, log.ID, DefaultTimeLogFields)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, log.ID, fetched.ID)
	assert.Equal(t, 30, fetched.TimeSpent)

	updated, err := c.TimeLogUpdate(ctx, log.ID, map[string]any{"time_spent": 60})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, log.ID, updated.ID, "update should return same time log ID")
	assert.Equal(t, 60, updated.TimeSpent)

	reFetched, _, err := c.TimeLog(ctx, log.ID, DefaultTimeLogFields)
	require.NoError(t, err)
	assert.Equal(t, 60, reFetched.TimeSpent, "update should persist")

	require.NoError(t, c.TimeLogDelete(ctx, log.ID))
	deleted = true

	_, _, err = c.TimeLog(ctx, log.ID, DefaultTimeLogFields)
	assert.Error(t, err, "deleted time log should no longer be readable")
}

func TestIntegration_TimeLogs_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	kwargs := map[string]any{
		"filter": []any{"parent.project_id", "==", projectID},
		"slice":  []int{0, 5},
	}

	logs, meta, err := c.TimeLogs(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, logs)

	for _, log := range logs {
		assert.True(t, strings.HasPrefix(log.ID, "CmfTimeTrackerHistory:"))
	}
}
