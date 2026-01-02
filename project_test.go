package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Project_Success_ReturnsProject(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfProject:123",
			"code": "PROJ-001",
			"name": "Test Project"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.get")
	}

	project, meta, err := client.Project(testCtx, "PROJ-001", nil)

	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "CmfProject:123", project.ID)
	assert.Equal(t, "PROJ-001", project.Code)
	assert.Equal(t, "Test Project", project.Name)
	assert.NotNil(t, meta)
}

func TestClient_Project_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Project not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	project, meta, err := client.Project(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Project not found")
	assert.Nil(t, project)
	assert.Nil(t, meta)
}

func TestClient_Project_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	project, meta, err := client.Project(testCtx, "PROJ-001", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, project)
	assert.Nil(t, meta)
}

func TestClient_ProjectsList_Success_ReturnsProjects(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfProject:1", "code": "PROJ-001", "name": "Project 1"},
			{"id": "CmfProject:2", "code": "PROJ-002", "name": "Project 2"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.list")
	}

	qb := NewQueryBuilder().From(EntityProject)
	projects, meta, err := client.ProjectsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "PROJ-001", projects[0].Code)
	assert.Equal(t, "PROJ-002", projects[1].Code)
	assert.NotNil(t, meta)
}

func TestClient_ProjectCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 25
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.count")
	}

	qb := NewQueryBuilder().From(EntityProject)
	count, err := client.ProjectCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 25, count)
}

func TestClient_Projects_Success_ReturnsProjects(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfProject:1", "code": "PROJ-001", "name": "Project 1"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.list")
	}

	projects, meta, err := client.Projects(testCtx, nil, nil)

	require.NoError(t, err)
	assert.Len(t, projects, 1)
	assert.NotNil(t, meta)
}

func TestClient_ProjectCreate_Success_ReturnsProject(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfProject:new-123",
			"code": "NEWPROJ",
			"name": "New Project"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.create")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		return assert.Contains(t, string(body), "New Project")
	}

	params := ProjectCreateParams{
		Code: "NEWPROJ",
		Name: "New Project",
	}
	project, err := client.ProjectCreate(testCtx, params)

	require.NoError(t, err)
	assert.Equal(t, "CmfProject:new-123", project.ID)
	assert.Equal(t, "New Project", project.Name)
}

func TestClient_ProjectUpdate_Success_ReturnsUpdatedProject(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfProject:123",
			"code": "PROJ-001",
			"name": "Updated Project"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.update")
	}

	updates := map[string]any{"name": "Updated Project"}
	project, err := client.ProjectUpdate(testCtx, "CmfProject:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Project", project.Name)
}

func TestClient_ProjectDelete_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.delete")
	}

	err := client.ProjectDelete(testCtx, "CmfProject:123")

	assert.NoError(t, err)
}

func TestClient_ProjectAddExecutor_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.add_executors")
	}

	err := client.ProjectAddExecutor(testCtx, "CmfProject:123", "CmfPerson:456")

	assert.NoError(t, err)
}

func TestClient_ProjectRemoveExecutor_Success_ReturnsNoError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.remove_executors")
	}

	err := client.ProjectRemoveExecutor(testCtx, "CmfProject:123", "CmfPerson:456")

	assert.NoError(t, err)
}
