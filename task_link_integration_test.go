package evateamclient

import (
	"context"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
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
