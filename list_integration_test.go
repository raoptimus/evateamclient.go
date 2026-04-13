package evateamclient

import (
	"context"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ListsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Limit(5)

	lists, meta, err := c.ListsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, lists, "project %s should have lists", integrationProjectCode)

	for _, l := range lists {
		assert.True(t, strings.HasPrefix(l.ID, "CmfList:"), "ID should start with CmfList:, got %s", l.ID)
		assert.NotEmpty(t, l.Name)
		assert.NotEmpty(t, l.Code)
		assert.Equal(t, projectID, l.ProjectID)
	}
}

func TestIntegration_ListsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultListListFields...).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Limit(5)

	lists, meta, err := c.ListsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, lists)

	for _, l := range lists {
		assert.True(t, strings.HasPrefix(l.ID, "CmfList:"))
		assert.NotEmpty(t, l.Code)
		assert.NotEmpty(t, l.Name)
		assert.Equal(t, projectID, l.ProjectID)
	}
}

func TestIntegration_List_ByCode(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a list code from the project first
	qb := NewQueryBuilder().
		Select(ListFieldID, ListFieldCode).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Limit(1)

	lists, _, err := c.ListsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, lists, "need at least one list")

	listCode := lists[0].Code
	require.NotEmpty(t, listCode)

	// Fetch single list by code with default fields
	l, meta, err := c.List(ctx, listCode, DefaultListFields)
	require.NoError(t, err)
	require.NotNil(t, l)
	require.NotNil(t, meta)

	assert.Equal(t, listCode, l.Code)
	assert.NotEmpty(t, l.ID)
	assert.NotEmpty(t, l.Name)
	assert.Equal(t, projectID, l.ProjectID)
}

func TestIntegration_List_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a list code
	qb := NewQueryBuilder().
		Select(ListFieldCode).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID}).
		Limit(1)

	lists, _, err := c.ListsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, lists)

	// Fetch with ** (all basic + relation fields)
	l, meta, err := c.List(ctx, lists[0].Code, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, l)
	require.NotNil(t, meta)

	assert.NotEmpty(t, l.ID)
	assert.NotEmpty(t, l.Name)
}

func TestIntegration_ListCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(ListFieldID).
		From(EntityList).
		Where(sq.Eq{ListFieldProjectID: projectID})

	count, err := c.ListCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "project %s should have lists", integrationProjectCode)
}

func TestIntegration_ProjectLists(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	lists, meta, err := c.ProjectLists(ctx, projectID, DefaultListListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, lists)

	for _, l := range lists {
		assert.True(t, strings.HasPrefix(l.ID, "CmfList:"))
		assert.Equal(t, projectID, l.ProjectID)
	}
}

func TestIntegration_ProjectSprints(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	sprints, meta, err := c.ProjectSprints(ctx, projectID, DefaultListListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(sprints) == 0 {
		t.Skip("no sprints found in project, skipping sprint type verification")
	}

	for _, s := range sprints {
		assert.True(t, strings.HasPrefix(s.ID, "CmfList:"))
		assert.Equal(t, projectID, s.ProjectID)
		assert.True(t, s.IsSprint(), "code %s should be a sprint (prefix SPR-)", s.Code)
	}
}

func TestIntegration_ProjectReleases(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	releases, meta, err := c.ProjectReleases(ctx, projectID, DefaultListListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(releases) == 0 {
		t.Skip("no releases found in project, skipping release type verification")
	}

	for _, r := range releases {
		assert.True(t, strings.HasPrefix(r.ID, "CmfList:"))
		assert.Equal(t, projectID, r.ProjectID)
		assert.True(t, r.IsRelease(), "code %s should be a release (prefix REL-)", r.Code)
	}
}

func TestIntegration_OpenProjectLists(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	lists, meta, err := c.OpenProjectLists(ctx, projectID, DefaultListListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(lists) == 0 {
		t.Skip("no open lists found in project")
	}

	for _, l := range lists {
		assert.True(t, strings.HasPrefix(l.ID, "CmfList:"))
		assert.Equal(t, projectID, l.ProjectID)
		assert.Equal(t, "OPEN", l.CacheStatusType, "list %s should have OPEN status", l.Code)
	}
}

func TestIntegration_Lists_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	kwargs := map[string]any{
		"filter": [][]any{
			{ListFieldProjectID, "==", projectID},
		},
		"slice": []int{0, 5},
	}

	lists, meta, err := c.Lists(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, lists)

	for _, l := range lists {
		assert.True(t, strings.HasPrefix(l.ID, "CmfList:"))
	}
}
