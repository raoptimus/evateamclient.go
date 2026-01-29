package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	evateamclient "github.com/raoptimus/evateamclient"
)

// ProjectTools provides MCP tool handlers for project operations.
type ProjectTools struct {
	client *evateamclient.Client
}

// NewProjectTools creates a new ProjectTools instance.
func NewProjectTools(client *evateamclient.Client) *ProjectTools {
	return &ProjectTools{client: client}
}

// ProjectListInput represents input for eva_project_list tool.
type ProjectListInput struct {
	QueryInput

	// Filter by system projects
	System *bool `json:"system,omitempty"`
}

// ProjectList returns a list of projects.
func (p *ProjectTools) ProjectList(ctx context.Context, input ProjectListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityProject, &input.QueryInput)
	if err != nil {
		return nil, WrapError("project_list", err)
	}

	// Add system filter if specified
	if input.System != nil {
		qb = qb.Where(sq.Eq{"system": *input.System})
	}

	projects, _, err := p.client.ProjectsList(ctx, qb)
	if err != nil {
		return nil, WrapError("project_list", err)
	}

	return &ListResult{
		Items:   projects,
		HasMore: len(projects) == input.Limit && input.Limit > 0,
	}, nil
}

// ProjectGetInput represents input for eva_project_get tool.
type ProjectGetInput struct {
	// Project code (e.g., "PROJ")
	Code string `json:"code,omitempty"`

	// Project ID (e.g., "CmfProject:uuid")
	ID string `json:"id,omitempty"`

	// Fields to return
	Fields []string `json:"fields,omitempty"`
}

// ProjectGet retrieves a single project by code or ID.
func (p *ProjectTools) ProjectGet(ctx context.Context, input ProjectGetInput) (any, error) {
	var qb *evateamclient.QueryBuilder

	if input.Code != "" {
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityProject).
			Where(sq.Eq{"code": input.Code}).
			Limit(1)
	} else if input.ID != "" {
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityProject).
			Where(sq.Eq{"id": input.ID}).
			Limit(1)
	} else {
		return nil, WrapError("project_get", ErrInvalidInput)
	}

	project, _, err := p.client.ProjectQuery(ctx, qb)
	if err != nil {
		return nil, WrapError("project_get", err)
	}

	return project, nil
}

// ProjectCreateInput represents input for eva_project_create tool.
type ProjectCreateInput struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Text       string   `json:"text,omitempty"`
	WorkflowID string   `json:"workflow_id,omitempty"`
	Executors  []string `json:"executors,omitempty"`
	Admins     []string `json:"admins,omitempty"`
}

// ProjectCreate creates a new project.
func (p *ProjectTools) ProjectCreate(ctx context.Context, input ProjectCreateInput) (any, error) {
	params := &evateamclient.ProjectCreateParams{
		Code:       input.Code,
		Name:       input.Name,
		Text:       input.Text,
		WorkflowID: input.WorkflowID,
		Executors:  input.Executors,
		Admins:     input.Admins,
	}

	project, err := p.client.ProjectCreate(ctx, params)
	if err != nil {
		return nil, WrapError("project_create", err)
	}

	return project, nil
}

// ProjectUpdateInput represents input for eva_project_update tool.
type ProjectUpdateInput struct {
	ID      string         `json:"id"`
	Updates map[string]any `json:"updates"`
}

// ProjectUpdate updates an existing project.
func (p *ProjectTools) ProjectUpdate(ctx context.Context, input ProjectUpdateInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("project_update", ErrInvalidInput)
	}

	project, err := p.client.ProjectUpdate(ctx, input.ID, input.Updates)
	if err != nil {
		return nil, WrapError("project_update", err)
	}

	return project, nil
}

// ProjectDeleteInput represents input for eva_project_delete tool.
type ProjectDeleteInput struct {
	ID string `json:"id"`
}

// ProjectDelete deletes a project.
func (p *ProjectTools) ProjectDelete(ctx context.Context, input ProjectDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("project_delete", ErrInvalidInput)
	}

	err := p.client.ProjectDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("project_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// ProjectAddExecutorInput represents input for eva_project_add_executor tool.
type ProjectAddExecutorInput struct {
	ProjectID string `json:"project_id"`
	PersonID  string `json:"person_id"`
}

// ProjectAddExecutor adds an executor to a project.
func (p *ProjectTools) ProjectAddExecutor(ctx context.Context, input ProjectAddExecutorInput) (any, error) {
	if input.ProjectID == "" || input.PersonID == "" {
		return nil, WrapError("project_add_executor", ErrInvalidInput)
	}

	err := p.client.ProjectAddExecutor(ctx, input.ProjectID, input.PersonID)
	if err != nil {
		return nil, WrapError("project_add_executor", err)
	}

	return map[string]bool{"success": true}, nil
}

// ProjectRemoveExecutorInput represents input for eva_project_remove_executor tool.
type ProjectRemoveExecutorInput struct {
	ProjectID string `json:"project_id"`
	PersonID  string `json:"person_id"`
}

// ProjectRemoveExecutor removes an executor from a project.
func (p *ProjectTools) ProjectRemoveExecutor(ctx context.Context, input ProjectRemoveExecutorInput) (any, error) {
	if input.ProjectID == "" || input.PersonID == "" {
		return nil, WrapError("project_remove_executor", ErrInvalidInput)
	}

	err := p.client.ProjectRemoveExecutor(ctx, input.ProjectID, input.PersonID)
	if err != nil {
		return nil, WrapError("project_remove_executor", err)
	}

	return map[string]bool{"success": true}, nil
}

// ProjectCountInput represents input for eva_project_count tool.
type ProjectCountInput struct {
	System *bool `json:"system,omitempty"`
}

// ProjectCount counts projects.
func (p *ProjectTools) ProjectCount(ctx context.Context, input ProjectCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityProject)

	if input.System != nil {
		qb = qb.Where(sq.Eq{"system": *input.System})
	}

	count, err := p.client.ProjectCount(ctx, qb)
	if err != nil {
		return nil, WrapError("project_count", err)
	}

	return &CountResult{Count: count}, nil
}
