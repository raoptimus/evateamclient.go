package evateamclient

import (
	"context"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Project_ByCode(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	project, meta, err := c.Project(ctx, integrationProjectCode, DefaultProjectFields)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotNil(t, meta)

	assert.Equal(t, integrationProjectCode, project.Code)
	assert.True(t, strings.HasPrefix(project.ID, "CmfProject:"), "ID should start with CmfProject:, got %s", project.ID)
	assert.NotEmpty(t, project.Name)
	assert.False(t, project.CMFCreatedAt.IsZero(), "CMFCreatedAt should not be zero")
	assert.False(t, project.CMFModifiedAt.IsZero(), "CMFModifiedAt should not be zero")
}

func TestIntegration_Project_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	project, meta, err := c.Project(ctx, integrationProjectCode, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, project)
	require.NotNil(t, meta)

	assert.NotEmpty(t, project.ID)
	assert.NotEmpty(t, project.Name)
	assert.Equal(t, integrationProjectCode, project.Code)

	// With ** fields, relation fields should be populated
	if len(project.Executors) > 0 {
		assert.NotEmpty(t, project.Executors[0].ID, "executor should have ID")
	}
	if len(project.Admins) > 0 {
		assert.NotEmpty(t, project.Admins[0].ID, "admin should have ID")
	}
}

func TestIntegration_ProjectQuery_WithBuilder(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(
			ProjectFieldID,
			ProjectFieldCode,
			ProjectFieldName,
			ProjectFieldCacheStatusType,
			ProjectFieldCMFCreatedAt,
			ProjectFieldCMFModifiedAt,
		).
		From(EntityProject).
		Where(sq.Eq{ProjectFieldCode: integrationProjectCode}).
		Limit(1)

	project, meta, err := c.ProjectQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotNil(t, meta)

	assert.True(t, strings.HasPrefix(project.ID, "CmfProject:"))
	assert.Equal(t, integrationProjectCode, project.Code)
	assert.NotEmpty(t, project.Name)
	assert.False(t, project.CMFCreatedAt.IsZero())
	assert.False(t, project.CMFModifiedAt.IsZero())
}

func TestIntegration_ProjectsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityProject).
		Limit(3)

	projects, meta, err := c.ProjectsList(ctx, qb)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, meta)
	require.NotEmpty(t, projects)

	for _, p := range projects {
		assert.True(t, strings.HasPrefix(p.ID, "CmfProject:"), "ID should start with CmfProject:, got %s", p.ID)
		assert.NotEmpty(t, p.Name)
		assert.False(t, p.CMFCreatedAt.IsZero(), "CMFCreatedAt should not be zero")
	}
}

func TestIntegration_ProjectsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultProjectListFields...).
		From(EntityProject).
		Limit(5)

	projects, meta, err := c.ProjectsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, projects)

	for _, p := range projects {
		assert.True(t, strings.HasPrefix(p.ID, "CmfProject:"))
		assert.NotEmpty(t, p.Code)
		assert.NotEmpty(t, p.Name)
	}
}

func TestIntegration_ProjectCount(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(ProjectFieldID).
		From(EntityProject)

	count, err := c.ProjectCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "should have at least one project")
}

func TestIntegration_Projects_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	kwargs := map[string]any{
		"slice": []int{0, 5},
	}

	projects, meta, err := c.Projects(ctx, nil, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, projects)

	for _, p := range projects {
		assert.True(t, strings.HasPrefix(p.ID, "CmfProject:"))
	}
}
