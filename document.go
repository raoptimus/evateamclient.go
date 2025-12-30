package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Document field constants for type-safe queries
const (
	// Core fields
	DocumentFieldID              = "id"
	DocumentFieldClassName       = "class_name"
	DocumentFieldCode            = "code"
	DocumentFieldName            = "name"
	DocumentFieldText            = "text"
	DocumentFieldProjectID       = "project_id"
	DocumentFieldParentID        = "parent_id"
	DocumentFieldCacheStatusType = "cache_status_type"

	// System
	DocumentFieldCmfCreatedAt  = "cmf_created_at"
	DocumentFieldCmfModifiedAt = "cmf_modified_at"
	DocumentFieldCmfOwnerID    = "cmf_owner_id"
	DocumentFieldCmfDeleted    = "cmf_deleted"
)

var (
	// DefaultDocumentFields - standard projection for single document queries
	DefaultDocumentFields = []string{
		DocumentFieldID,
		DocumentFieldClassName,
		DocumentFieldCode,
		DocumentFieldName,
		DocumentFieldText,
		DocumentFieldProjectID,
		DocumentFieldCacheStatusType,
		DocumentFieldCmfCreatedAt,
		DocumentFieldCmfModifiedAt,
	}

	// DefaultDocumentListFields - optimized for LIST queries
	DefaultDocumentListFields = []string{
		DocumentFieldID,
		DocumentFieldCode,
		DocumentFieldName,
		DocumentFieldProjectID,
		DocumentFieldCacheStatusType,
		DocumentFieldCmfCreatedAt,
	}
)

// Document retrieves a single document by code
// Example:
//
//	doc, meta, err := client.Document(ctx, "DOC-123", nil)
func (c *Client) Document(
	ctx context.Context,
	docCode string,
	fields []string,
) (*models.Document, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldCode: docCode}).
		Limit(1)

	return c.DocumentQuery(ctx, qb)
}

// DocumentQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name", "text").
//	  From(evateamclient.EntityDocument).
//	  Where(sq.Eq{"code": "DOC-123"})
//	doc, meta, err := client.DocumentQuery(ctx, qb)
func (c *Client) DocumentQuery(ctx context.Context, qb *QueryBuilder) (*models.Document, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultDocumentFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// DocumentsList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityDocument).
//	  Where(sq.Eq{"project_id": "Project:uuid"}).
//	  OrderBy("-cmf_created_at").
//	  Offset(0).Limit(100)
//	docs, meta, err := client.DocumentsList(ctx, qb)
func (c *Client) DocumentsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Document, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultDocumentListFields
	}

	method, err := qb.ToMethod()
	if err != nil {
		return nil, nil, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// DocumentCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityDocument).
//	  Where(sq.Eq{"project_id": "Project:uuid"})
//	count, err := client.DocumentCount(ctx, qb)
func (c *Client) DocumentCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  int    `json:"result"`
	}

	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// ProjectDocuments retrieves ALL documents in project
// Example:
//
//	docs, meta, err := client.ProjectDocuments(ctx, "Project:uuid", nil)
func (c *Client) ProjectDocuments(
	ctx context.Context,
	projectID string,
	fields []string,
) ([]models.Document, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		OrderBy("-" + DocumentFieldCmfCreatedAt)

	return c.DocumentsList(ctx, qb)
}

// CRUD Operations

// DocumentCreateParams contains parameters for creating a new document
type DocumentCreateParams struct {
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	Text      string `json:"text,omitempty"`
	ParentID  string `json:"parent_id,omitempty"`
}

// DocumentCreate creates a new document
// Example:
//
//	params := evateamclient.DocumentCreateParams{
//	  Name:      "New Document",
//	  ProjectID: "Project:uuid",
//	  Text:      "Document content",
//	}
//	doc, err := client.DocumentCreate(ctx, params)
func (c *Client) DocumentCreate(
	ctx context.Context,
	params DocumentCreateParams,
) (*models.Document, error) {
	kwargs := map[string]any{
		"name":       params.Name,
		"project_id": params.ProjectID,
	}

	if params.Text != "" {
		kwargs["text"] = params.Text
	}
	if params.ParentID != "" {
		kwargs["parent_id"] = params.ParentID
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// DocumentUpdate updates an existing document
// Example:
//
//	updates := map[string]any{
//	  "name": "Updated Document Name",
//	  "text": "Updated content",
//	}
//	doc, err := client.DocumentUpdate(ctx, "CmfDocument:uuid", updates)
func (c *Client) DocumentUpdate(
	ctx context.Context,
	docID string,
	updates map[string]any,
) (*models.Document, error) {
	kwargs := map[string]any{
		"id": docID,
	}
	for k, v := range updates {
		kwargs[k] = v
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.update",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return &resp.Result, nil
}

// DocumentDelete deletes a document by ID
// Example:
//
//	err := client.DocumentDelete(ctx, "CmfDocument:uuid")
func (c *Client) DocumentDelete(
	ctx context.Context,
	docID string,
) error {
	kwargs := map[string]any{
		"id": docID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.delete",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  bool   `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// Backward compatible methods (using old API)

// Documents retrieves documents with custom filters (backward compatible, deprecated)
// Recommended: use DocumentsList with NewQueryBuilder() instead
func (c *Client) Documents(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Document, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultDocumentListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
