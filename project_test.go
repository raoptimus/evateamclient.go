/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	encjson "encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/imroc/req/v3"
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
	project, err := client.ProjectCreate(testCtx, &params)

	require.NoError(t, err)
	assert.Equal(t, "CmfProject:new-123", project.ID)
	assert.Equal(t, "New Project", project.Name)
}

func TestClient_ProjectCreate_ResultIsIDString_FetchesCreatedProject(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfProject:new-123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfProject:new-123", "code": "NEWPROJ", "name": "New Project"}
		}`),
	}

	params := ProjectCreateParams{Code: "NEWPROJ", Name: "New Project"}
	project, err := client.ProjectCreate(testCtx, &params)

	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "CmfProject:new-123", project.ID)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected create + follow-up get")
}

// TestClient_ProjectCreate_FollowUpGet_FiltersByIDOrCode covers both forms a
// bare create-result string can take: an ID ("CmfProject:uuid", which
// carries a ":" class-name prefix) or a code ("NEWPROJ", which doesn't).
// Filtering the follow-up .get by the wrong field would find nothing and
// surface as a spurious error, prompting callers to retry and create dupes.
func TestClient_ProjectCreate_FollowUpGet_FiltersByIDOrCode(t *testing.T) {
	tests := []struct {
		name        string
		resultValue string
		wantField   string
	}{
		{"ID form filters by id", "CmfProject:new-123", ProjectFieldID},
		{"code form filters by code", "NEWPROJ", ProjectFieldCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClientWithSequentialMock(t)

			mockHTTP.responses = []*req.Response{
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"`+tt.resultValue+`"}`),
				mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":{"id":"CmfProject:new-123","code":"NEWPROJ"}}`),
			}

			var followUpFilter []any
			mockHTTP.bodyCheck = func(body []byte) bool {
				var parsed struct {
					Method string         `json:"method"`
					Kwargs map[string]any `json:"kwargs"`
				}
				if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
					return false
				}
				if parsed.Method != "CmfProject.get" {
					return true
				}
				filter, ok := parsed.Kwargs["filter"].([]any)
				if !assert.True(t, ok, "expected filter kwarg on follow-up .get") {
					return false
				}
				followUpFilter = filter
				return true
			}

			params := ProjectCreateParams{Code: "NEWPROJ", Name: "New Project"}
			project, err := client.ProjectCreate(testCtx, &params)

			require.NoError(t, err)
			require.NotNil(t, project)
			require.NotEmpty(t, followUpFilter, "follow-up .get should have been called with a filter")
			assert.Equal(t, tt.wantField, followUpFilter[0])
		})
	}
}

// TestClient_ProjectCreate_EmptyResult_ReturnsError guards against a silent
// failure (empty/null/false result) being mistaken for a zero-value project.
func TestClient_ProjectCreate_EmptyResult_ReturnsError(t *testing.T) {
	tests := []struct {
		name   string
		result string
	}{
		{"null", `null`},
		{"empty string", `""`},
		{"false", `false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockHTTP := newTestClient(t)
			mockHTTP.response = mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":`+tt.result+`}`)

			params := ProjectCreateParams{Code: "NEWPROJ", Name: "New Project"}
			project, err := client.ProjectCreate(testCtx, &params)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "CmfProject.create returned empty result")
			assert.Nil(t, project)
		})
	}
}

// TestClient_ProjectUpdate_Success_SendsIDInArgs is the regression test for
// the EVA update convention: the id belongs in Args, not Kwargs (see TaskUpdate).
func TestClient_ProjectUpdate_Success_SendsIDInArgs(t *testing.T) {
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
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfProject:123"}, parsed.Args) &&
			assert.Equal(t, map[string]any{"name": "Updated Project"}, parsed.Kwargs)
	}

	updates := map[string]any{"name": "Updated Project"}
	project, err := client.ProjectUpdate(testCtx, "CmfProject:123", updates)

	require.NoError(t, err)
	assert.Equal(t, "Updated Project", project.Name)
}

func TestClient_ProjectUpdate_ResultIsIDString_FetchesUpdatedProject(t *testing.T) {
	client, mockHTTP := newTestClientWithSequentialMock(t)

	mockHTTP.responses = []*req.Response{
		mockResponse(http.StatusOK, `{"jsonrpc":"2.2","result":"CmfProject:123"}`),
		mockResponse(http.StatusOK, `{
			"jsonrpc": "2.2",
			"result": {"id": "CmfProject:123", "code": "PROJ-001", "name": "Updated Project"}
		}`),
	}

	updates := map[string]any{"name": "Updated Project"}
	project, err := client.ProjectUpdate(testCtx, "CmfProject:123", updates)

	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "Updated Project", project.Name)
	assert.Equal(t, 2, mockHTTP.callIdx, "expected update + follow-up get")
}

func TestClient_ProjectUpdate_EmptyProjectID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	project, err := client.ProjectUpdate(testCtx, "", map[string]any{"name": "x"})

	require.Error(t, err)
	assert.Nil(t, project)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
}

// TestClient_ProjectDelete_Success_SendsIDInArgs is the regression test for
// the EVA delete convention: the id belongs in Args, not Kwargs (see TaskDelete).
func TestClient_ProjectDelete_Success_SendsIDInArgs(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": true
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfProject.delete")
	}
	mockHTTP.bodyCheck = func(body []byte) bool {
		var parsed struct {
			Args   []any          `json:"args"`
			Kwargs map[string]any `json:"kwargs"`
		}
		if err := encjson.Unmarshal(body, &parsed); !assert.NoError(t, err) {
			return false
		}
		return assert.Equal(t, []any{"CmfProject:123"}, parsed.Args) &&
			assert.Nil(t, parsed.Kwargs)
	}

	err := client.ProjectDelete(testCtx, "CmfProject:123")

	assert.NoError(t, err)
}

func TestClient_ProjectDelete_EmptyProjectID_ReturnsErrorWithoutRequest(t *testing.T) {
	client, mockHTTP := newTestClient(t)
	mockHTTP.err = errors.New("must not be called")

	err := client.ProjectDelete(testCtx, "")

	assert.Error(t, err)
	assert.Equal(t, 0, mockHTTP.calls, "validation must fail before any HTTP request")
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
