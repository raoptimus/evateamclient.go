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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ProjectTasksCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	count, meta, err := c.ProjectTasksCount(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, meta)
	assert.Greater(t, count, int64(0), "project %s should have tasks", integrationProjectCode)
}

func TestIntegration_SprintStats(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a sprint code from the project
	sprints, _, err := c.ProjectSprints(ctx, projectID, []string{ListFieldID, ListFieldCode})
	require.NoError(t, err)

	if len(sprints) == 0 {
		t.Skip("no sprints found in project, skipping sprint stats test")
	}

	sprintCode := sprints[0].Code
	require.NotEmpty(t, sprintCode)

	stats, err := c.SprintStats(ctx, sprintCode)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, sprintCode, stats.SprintID)
	assert.GreaterOrEqual(t, stats.TotalTasks, 0)
	assert.NotNil(t, stats.TasksByStatus)
}

func TestIntegration_ProjectStats(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	stats, _, err := c.ProjectStats(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, projectID, stats.ProjectID)
	assert.Greater(t, stats.TotalTasks, 0, "project %s should have tasks", integrationProjectCode)
	assert.GreaterOrEqual(t, stats.OpenTasks, 0)
	assert.GreaterOrEqual(t, stats.ActiveSprints, 0)
	assert.GreaterOrEqual(t, stats.TotalUsers, 0)
}
