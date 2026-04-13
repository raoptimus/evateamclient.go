package evateamclient

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_ProjectEpics(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	epics, meta, err := c.ProjectEpics(ctx, projectID, DefaultEpicListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	for _, epic := range epics {
		assert.True(t, strings.HasPrefix(epic.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", epic.ID)
		assert.NotEmpty(t, epic.Code)
		assert.NotEmpty(t, epic.Name)
		assert.Equal(t, projectID, epic.ProjectID)
	}
}

func TestIntegration_ProjectEpics_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	epics, meta, err := c.ProjectEpics(ctx, projectID, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, meta)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	for _, epic := range epics {
		assert.True(t, strings.HasPrefix(epic.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", epic.ID)
		assert.NotEmpty(t, epic.Name)
		assert.Equal(t, projectID, epic.ProjectID)
	}
}

func TestIntegration_Epic_ByCode(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get first epic code from ProjectEpics
	epics, _, err := c.ProjectEpics(ctx, projectID, DefaultEpicListFields)
	require.NoError(t, err)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	epicCode := epics[0].Code
	require.NotEmpty(t, epicCode)

	// Fetch single epic by code
	epic, meta, err := c.Epic(ctx, epicCode, DefaultEpicFields)
	require.NoError(t, err)
	require.NotNil(t, epic)
	require.NotNil(t, meta)

	assert.Equal(t, epicCode, epic.Code)
	assert.True(t, strings.HasPrefix(epic.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", epic.ID)
	assert.NotEmpty(t, epic.Name)
	assert.Equal(t, projectID, epic.ProjectID)
}

func TestIntegration_Epic_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get first epic code
	epics, _, err := c.ProjectEpics(ctx, projectID, DefaultEpicListFields)
	require.NoError(t, err)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	epicCode := epics[0].Code
	require.NotEmpty(t, epicCode)

	// Fetch with ** (all basic + relation fields)
	epic, meta, err := c.Epic(ctx, epicCode, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, epic)
	require.NotNil(t, meta)

	assert.True(t, strings.HasPrefix(epic.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", epic.ID)
	assert.NotEmpty(t, epic.Name)
}

func TestIntegration_EpicByID(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get first epic to obtain its ID
	epics, _, err := c.ProjectEpics(ctx, projectID, DefaultEpicListFields)
	require.NoError(t, err)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	epicID := epics[0].ID
	require.NotEmpty(t, epicID)
	require.True(t, strings.HasPrefix(epicID, "CmfTask:"))

	// Fetch epic by ID
	epic, meta, err := c.EpicByID(ctx, epicID, DefaultEpicFields)
	require.NoError(t, err)
	require.NotNil(t, epic)
	require.NotNil(t, meta)

	assert.Equal(t, epicID, epic.ID)
	assert.NotEmpty(t, epic.Code)
	assert.NotEmpty(t, epic.Name)
	assert.Equal(t, projectID, epic.ProjectID)
}

func TestIntegration_EpicTasks(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get first epic to obtain its ID
	epics, _, err := c.ProjectEpics(ctx, projectID, DefaultEpicListFields)
	require.NoError(t, err)

	if len(epics) == 0 {
		t.Skip("no epics found in project, skipping")
	}

	epicID := epics[0].ID
	require.NotEmpty(t, epicID)

	// Get tasks within the epic
	tasks, meta, err := c.EpicTasks(ctx, epicID, DefaultTaskListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)

	// Epic may or may not have tasks, just verify no errors and proper format
	for _, task := range tasks {
		assert.True(t, strings.HasPrefix(task.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", task.ID)
		assert.NotEmpty(t, task.Code)
	}
}

func TestIntegration_Epics_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	kwargs := map[string]any{
		"filter": [][]any{
			{TaskFieldProjectID, "==", projectID},
			{"logic_type.code", "==", LogicTypeEpic},
		},
		"slice": []int{0, 5},
	}

	epics, meta, err := c.Epics(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)

	// May be empty if project has no epics
	for _, epic := range epics {
		assert.True(t, strings.HasPrefix(epic.ID, "CmfTask:"), "ID should start with CmfTask:, got %s", epic.ID)
	}
}
