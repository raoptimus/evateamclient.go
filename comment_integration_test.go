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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getIntegrationTaskCode fetches the first task code from the integration project.
func getIntegrationTaskCode(t *testing.T, c *Client, projectID string) string {
	t.Helper()

	ctx := context.Background()
	qb := NewQueryBuilder().
		Select(TaskFieldID, TaskFieldCode).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "project should have at least one task")
	require.NotEmpty(t, tasks[0].Code)

	return tasks[0].Code
}

// getIntegrationTaskID fetches the first task ID from the integration project.
func getIntegrationTaskID(t *testing.T, c *Client, projectID string) string {
	t.Helper()

	ctx := context.Background()
	qb := NewQueryBuilder().
		Select(TaskFieldID).
		From(EntityTask).
		Where(sq.Eq{TaskFieldProjectID: projectID}).
		Limit(1)

	tasks, _, err := c.TasksList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "project should have at least one task")
	require.NotEmpty(t, tasks[0].ID)

	return tasks[0].ID
}

func TestIntegration_CommentsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldParentID: taskID}).
		OrderBy("-" + CommentFieldCmfCreatedAt).
		Limit(5)

	comments, meta, err := c.CommentsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	for _, comment := range comments {
		assert.True(t, strings.HasPrefix(comment.ID, "CmfComment:"),
			"ID should start with CmfComment:, got %s", comment.ID)
		assert.NotEmpty(t, comment.ParentID)
		assert.Equal(t, taskID, comment.ParentID)
	}
}

func TestIntegration_CommentsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)

	qb := NewQueryBuilder().
		Select(DefaultCommentListFields...).
		From(EntityComment).
		Where(sq.Eq{CommentFieldParentID: taskID}).
		Limit(5)

	comments, meta, err := c.CommentsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	for _, comment := range comments {
		assert.True(t, strings.HasPrefix(comment.ID, "CmfComment:"))
		assert.NotEmpty(t, comment.AuthorID)
		assert.False(t, comment.CreatedAt.IsZero(), "CreatedAt should be parsed")
	}
}

func TestIntegration_Comment_ByID(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a comment ID from a task
	taskID := getIntegrationTaskID(t, c, projectID)

	qb := NewQueryBuilder().
		Select(CommentFieldID).
		From(EntityComment).
		Where(sq.Eq{CommentFieldParentID: taskID}).
		Limit(1)

	comments, _, err := c.CommentsList(ctx, qb)
	require.NoError(t, err)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	commentID := comments[0].ID
	require.NotEmpty(t, commentID)

	// Fetch single comment by ID
	comment, meta, err := c.Comment(ctx, commentID, DefaultCommentFields)
	require.NoError(t, err)
	require.NotNil(t, comment)
	require.NotNil(t, meta)

	assert.Equal(t, commentID, comment.ID)
	assert.NotEmpty(t, comment.AuthorID)
	assert.False(t, comment.CreatedAt.IsZero(), "CreatedAt should be parsed as time.Time")
}

func TestIntegration_Comment_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)

	qb := NewQueryBuilder().
		Select(CommentFieldID).
		From(EntityComment).
		Where(sq.Eq{CommentFieldParentID: taskID}).
		Limit(1)

	comments, _, err := c.CommentsList(ctx, qb)
	require.NoError(t, err)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	// Fetch with ** (all basic + relation fields)
	comment, meta, err := c.Comment(ctx, comments[0].ID, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, comment)
	require.NotNil(t, meta)

	assert.NotEmpty(t, comment.ID)
	assert.True(t, strings.HasPrefix(comment.ID, "CmfComment:"))
}

func TestIntegration_CommentCount(t *testing.T) {
	t.Skip("EVA API does not support CmfComment.count method")
}

func TestIntegration_TaskComments(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskCode := getIntegrationTaskCode(t, c, projectID)

	comments, meta, err := c.TaskComments(ctx, taskCode, DefaultCommentListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	for _, comment := range comments {
		assert.True(t, strings.HasPrefix(comment.ID, "CmfComment:"),
			"ID should start with CmfComment:, got %s", comment.ID)
		assert.NotEmpty(t, comment.AuthorID)
		assert.False(t, comment.CreatedAt.IsZero(), "CreatedAt should be parsed")
	}
}

// TestIntegration_CommentCreate_UpdateDelete_FullCycle is the red-green
// regression test for SPEC-02: CommentCreate/CommentUpdate/CommentDelete
// must actually write, not silently no-op (id-in-kwargs bug), fail to parse
// a bare-string result (two-phase parsing bug), or call the nonexistent
// "Comment" class instead of "CmfComment".
func TestIntegration_CommentCreate_UpdateDelete_FullCycle(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)
	suffix := time.Now().UnixNano()

	comment, err := c.CommentCreate(ctx, taskID, fmt.Sprintf("[TEST] comment %d", suffix))
	require.NoError(t, err)
	require.NotNil(t, comment)
	require.NotEmpty(t, comment.ID, "CommentCreate must return a created comment, not a silent no-op")

	deleted := false
	t.Cleanup(func() {
		if !deleted {
			_ = c.CommentDelete(context.Background(), comment.ID)
		}
	})

	fetched, _, err := c.Comment(ctx, comment.ID, DefaultCommentFields)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, comment.ID, fetched.ID)

	updatedText := fmt.Sprintf("[TEST] comment UPDATED %d", suffix)
	updated, err := c.CommentUpdate(ctx, comment.ID, updatedText)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, comment.ID, updated.ID, "update should return same comment ID")
	assert.Equal(t, updatedText, updated.Text)

	reFetched, _, err := c.Comment(ctx, comment.ID, DefaultCommentFields)
	require.NoError(t, err)
	assert.Equal(t, updatedText, reFetched.Text, "update should persist")

	require.NoError(t, c.CommentDelete(ctx, comment.ID))
	deleted = true

	_, _, err = c.Comment(ctx, comment.ID, DefaultCommentFields)
	assert.Error(t, err, "deleted comment should no longer be readable")
}

func TestIntegration_Comments_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	taskID := getIntegrationTaskID(t, c, projectID)

	kwargs := map[string]any{
		"filter": [][]any{
			{CommentFieldParentID, "==", taskID},
		},
		"slice": []int{0, 5},
	}

	comments, meta, err := c.Comments(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(comments) == 0 {
		t.Skip("no comments found for task, skipping")
	}

	for _, comment := range comments {
		assert.True(t, strings.HasPrefix(comment.ID, "CmfComment:"))
	}
}
