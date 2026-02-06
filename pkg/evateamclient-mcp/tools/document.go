package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

// DocumentTools provides MCP tool handlers for document operations.
type DocumentTools struct {
	client *evateamclient.Client
}

// NewDocumentTools creates a new DocumentTools instance.
func NewDocumentTools(client *evateamclient.Client) *DocumentTools {
	return &DocumentTools{client: client}
}

// DocumentListInput represents input for eva_document_list tool.
type DocumentListInput struct {
	QueryInput
	ProjectID string `json:"project_id,omitempty"`
}

// DocumentList returns a list of documents.
func (d *DocumentTools) DocumentList(ctx context.Context, input *DocumentListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityDocument, &input.QueryInput)
	if err != nil {
		return nil, WrapError("document_list", err)
	}

	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}

	docs, _, err := d.client.DocumentsList(ctx, qb)
	if err != nil {
		return nil, WrapError("document_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(docs),
		HasMore: len(docs) == input.Limit && input.Limit > 0,
	}, nil
}

// DocumentGetInput represents input for eva_document_get tool.
type DocumentGetInput struct {
	Code   string   `json:"code,omitempty"`
	ID     string   `json:"id,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

// DocumentGet retrieves a single document.
func (d *DocumentTools) DocumentGet(ctx context.Context, input DocumentGetInput) (any, error) {
	var qb *evateamclient.QueryBuilder

	switch {
	case input.Code != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityDocument).
			Where(sq.Eq{"code": input.Code}).
			Limit(1)
	case input.ID != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityDocument).
			Where(sq.Eq{"id": input.ID}).
			Limit(1)
	default:
		return nil, WrapError("document_get", ErrInvalidInput)
	}

	doc, _, err := d.client.DocumentQuery(ctx, qb)
	if err != nil {
		return nil, WrapError("document_get", err)
	}

	return doc, nil
}

// DocumentCreateInput represents input for eva_document_create tool.
type DocumentCreateInput struct {
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	Text      string `json:"text,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
}

// DocumentCreate creates a new document.
func (d *DocumentTools) DocumentCreate(ctx context.Context, input DocumentCreateInput) (any, error) {
	params := evateamclient.DocumentCreateParams{
		Name:      input.Name,
		ProjectID: input.ProjectID,
		Text:      input.Text,
		ParentID:  input.ParentID,
	}

	doc, err := d.client.DocumentCreate(ctx, params)
	if err != nil {
		return nil, WrapError("document_create", err)
	}

	return doc, nil
}

// DocumentUpdateInput represents input for eva_document_update tool.
type DocumentUpdateInput struct {
	ID      string         `json:"id"`
	Updates map[string]any `json:"updates"`
}

// DocumentUpdate updates an existing document.
func (d *DocumentTools) DocumentUpdate(ctx context.Context, input DocumentUpdateInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("document_update", ErrInvalidInput)
	}

	doc, err := d.client.DocumentUpdate(ctx, input.ID, input.Updates)
	if err != nil {
		return nil, WrapError("document_update", err)
	}

	return doc, nil
}

// DocumentDeleteInput represents input for eva_document_delete tool.
type DocumentDeleteInput struct {
	ID string `json:"id"`
}

// DocumentDelete deletes a document.
func (d *DocumentTools) DocumentDelete(ctx context.Context, input DocumentDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("document_delete", ErrInvalidInput)
	}

	err := d.client.DocumentDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("document_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// DocumentCountInput represents input for eva_document_count tool.
type DocumentCountInput struct {
	ProjectID string `json:"project_id,omitempty"`
}

// DocumentCount counts documents.
func (d *DocumentTools) DocumentCount(ctx context.Context, input DocumentCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityDocument)

	if input.ProjectID != "" {
		qb = qb.Where(sq.Eq{"project_id": input.ProjectID})
	}

	count, err := d.client.DocumentCount(ctx, qb)
	if err != nil {
		return nil, WrapError("document_count", err)
	}

	return &CountResult{Count: count}, nil
}
