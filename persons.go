package evateamclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultPersonFields for person queries.
var DefaultPersonFields = []string{
	"id", "name", "email", "login", "active", "position", "department",
}

// ProjectPersons retrieves ALL users assigned to project.
func (c *Client) ProjectPersons(ctx context.Context, projectCode string, fields []string) ([]models.CmfPerson, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultPersonFields
	}

	kwargs := map[string]any{
		"filter": []any{"projects", "contains", fmt.Sprintf("CmfProject:%s", projectCode)},
		"fields": fields,
	}

	return c.Persons(ctx, kwargs)
}

// Person retrieves single user by ID or email/login.
func (c *Client) Person(ctx context.Context, userID string, fields []string) (*models.CmfPerson, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultPersonFields
	}

	kwargs := map[string]any{
		"filter": []any{"id", "==", userID},
		"fields": fields,
		"slice":  []int{0, 1},
	}

	users, meta, err := c.Persons(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}

	if len(users) == 0 {
		return nil, meta, nil
	}

	return &users[0], meta, nil
}

// Persons retrieves users with custom filters.
func (c *Client) Persons(ctx context.Context, kwargs map[string]any) ([]models.CmfPerson, *models.CmfMeta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfPerson.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfPersonListResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfPerson.list", reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// ProjectTaskExecutors retrieves all unique executors/responsibles for project tasks.
func (c *Client) ProjectTaskExecutors(ctx context.Context, projectCode string) ([]models.CmfPerson, *models.CmfMeta, error) {
	// Get tasks with responsible/executors fields
	tasks, _, err := c.ProjectTasks(ctx, projectCode, []string{"responsible", "executors"})
	if err != nil {
		return nil, nil, err
	}

	// Extract unique person IDs
	personIDs := make(map[string]bool)
	for _, task := range tasks {
		if task.Responsible != "" {
			personIDs[task.Responsible] = true
		}
		// Parse executors array if available
	}

	// Fetch persons by IDs
	idSlice := make([]string, 0, len(personIDs))
	for id := range personIDs {
		idSlice = append(idSlice, id)
	}

	filter := make([]any, 0, len(idSlice))
	for _, id := range idSlice {
		filter = append(filter, []any{"id", "==", id})
	}

	kwargs := map[string]any{
		"filter": filter,
		"fields": DefaultPersonFields,
	}

	return c.Persons(ctx, kwargs)
}
