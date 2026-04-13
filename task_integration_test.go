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

func TestIntegration_TasksList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		OrderBy("-" + TaskFieldCmfCreatedAt).
		Limit(5)

	tasks, meta, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, tasks, "project %s should have tasks", integrationProjectCode)

	for _, task := range tasks {
		assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", task.ID)
		assert.NotEmpty(t, task.Name)
		assert.Equal(t, projectID, task.ProjectID)
	}
}

func TestIntegration_TasksList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultTaskListFields...).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(5)

	tasks, meta, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, tasks)

	for _, task := range tasks {
		assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"))
		assert.NotEmpty(t, task.Code)
		assert.NotEmpty(t, task.Name)
		assert.Equal(t, projectID, task.ProjectID)
	}
}

func TestIntegration_Task_ByCode(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a task code from the list first
	qb := NewQueryBuilder().
		Select(TaskFieldID, TaskFieldCode).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "need at least one task")

	taskCode := tasks[0].Code
	require.NotEmpty(t, taskCode)

	// Fetch single task by code with fields including time fields
	fields := append(DefaultTaskFields, TaskFieldCmfCreatedAt, TaskFieldCmfModifiedAt)
	task, meta, err := c.Task(ctx, taskCode, fields)
	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotNil(t, meta)

	assert.Equal(t, taskCode, task.Code)
	assert.NotEmpty(t, task.ID)
	assert.NotEmpty(t, task.Name)
	assert.False(t, task.CmfCreatedAt.IsZero(), "CmfCreatedAt should be parsed as time.Time")
	assert.False(t, task.CmfModifiedAt.IsZero(), "CmfModifiedAt should be parsed as time.Time")
}

func TestIntegration_Task_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a task code
	qb := NewQueryBuilder().
		Select(TaskFieldCode).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)

	// Fetch with ** (all basic + relation fields)
	task, meta, err := c.Task(ctx, tasks[0].Code, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, task)
	require.NotNil(t, meta)

	assert.NotEmpty(t, task.ID)
	assert.NotEmpty(t, task.Name)
}

func TestIntegration_TaskQuery_WithBuilder(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(
			TaskFieldID,
			TaskFieldCode,
			TaskFieldName,
			TaskFieldText,
			TaskFieldProjectID,
			TaskFieldCacheStatusType,
			TaskFieldPriority,
			TaskFieldCmfCreatedAt,
			TaskFieldCmfModifiedAt,
		).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	task, meta, err := c.TaskQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, task)
	require.NotNil(t, meta)

	assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"))
	assert.NotEmpty(t, task.Code)
	assert.NotEmpty(t, task.Name)
	assert.Equal(t, projectID, task.ProjectID)
	assert.False(t, task.CmfCreatedAt.IsZero())
	assert.False(t, task.CmfModifiedAt.IsZero())
}

func TestIntegration_TaskCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(TaskFieldID).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID})

	count, err := c.TaskCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "project %s should have tasks", integrationProjectCode)
}

func TestIntegration_ProjectTasks(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	tasks, meta, err := c.ProjectTasks(ctx, projectID, DefaultTaskListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, tasks)

	for _, task := range tasks {
		assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"))
		assert.Equal(t, projectID, task.ParentID)
	}
}

func TestIntegration_Tasks_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	kwargs := map[string]any{
		"filter": [][]any{
			{TaskFieldProjectID, "==", projectID},
		},
		"slice": []int{0, 5},
	}

	tasks, meta, err := c.Tasks(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, tasks)

	for _, task := range tasks {
		assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"))
	}
}
