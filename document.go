/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"context"
	encjson "encoding/json"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/raoptimus/evateamclient.go/models"
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

	// CmfDocument.create kwargs (OAS additionalProperties: false; differ from
	// the read/filter field names above).
	documentCreateParent     = "parent"      // project ID
	documentCreateTreeParent = "tree_parent" // page-tree parent document
	documentCreateTextDraft  = "text_draft"  // draft text; visible `text` appears after do_publish
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

	method, err := qb.ToMethod(false)
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

// DocumentCreate creates a new document. Text (if given) is stored as a
// draft: the document is only visibly published once the client also calls
// CmfDocument.do_publish, which this method does automatically when Text is
// non-empty.
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
		DocumentFieldName:    params.Name,
		documentCreateParent: params.ProjectID,
	}

	if params.Text != "" {
		kwargs[documentCreateTextDraft] = params.Text
	}
	if params.ParentID != "" {
		kwargs[documentCreateTreeParent] = params.ParentID
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.create",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string             `json:"jsonrpc"`
		Result  encjson.RawMessage `json:"result"`
	}
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	doc, err := parseWriteResult(ctx, resp.Result, "CmfDocument.create", c.documentByID, documentHasEmptyID)
	if err != nil {
		return nil, err
	}

	if params.Text != "" {
		if pubErr := c.DocumentPublish(ctx, doc.ID); pubErr != nil {
			return doc, errors.WithMessagef(pubErr, "document %s created, publish failed; do not retry create", doc.ID)
		}
	}

	return doc, nil
}

// DocumentPublish promotes a document's draft text (text_draft) to the
// visible `text` field via CmfDocument.do_publish.
// Example:
//
//	err := client.DocumentPublish(ctx, "CmfDocument:uuid")
func (c *Client) DocumentPublish(ctx context.Context, docID string) error {
	if docID == "" {
		return errors.New("docID is required")
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.do_publish",
		CallID:  newCallID(),
		Args:    []any{docID},
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  any    `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// documentByID fetches a document by ID or code, for the two-phase
// create/update follow-up `.get` when CmfDocument.create/update returns a
// bare string. A ":" marks the class-name-prefixed ID form; otherwise it's a
// code (mirrors fetchTaskLinkByIDOrCode in task_link.go).
func (c *Client) documentByID(ctx context.Context, idOrCode string) (*models.Document, error) {
	field := DocumentFieldCode
	if strings.Contains(idOrCode, ":") {
		field = DocumentFieldID
	}

	qb := NewQueryBuilder().
		Select(DefaultDocumentFields...).
		From(EntityDocument).
		Where(sq.Eq{field: idOrCode}).
		Limit(1)

	doc, _, err := c.DocumentQuery(ctx, qb)
	return doc, err
}

func documentHasEmptyID(doc *models.Document) bool {
	return doc == nil || doc.ID == ""
}

// DocumentUpdate updates an existing document. A `text` key in updates is
// sent as text_draft (OAS: CmfDocument.update has no plain `text` kwarg) and,
// once applied, published automatically via DocumentPublish so the change
// becomes visible; pass `text_draft` directly instead to save a draft without
// publishing.
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
	if docID == "" {
		return nil, errors.New("docID is required")
	}

	text, hasText := updates[DocumentFieldText]
	kwargs := updates
	if hasText {
		kwargs = make(map[string]any, len(updates))
		for k, v := range updates {
			kwargs[k] = v
		}
		delete(kwargs, DocumentFieldText)
		kwargs[documentCreateTextDraft] = text
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.update",
		CallID:  newCallID(),
		Args:    []any{docID},
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string             `json:"jsonrpc"`
		Result  encjson.RawMessage `json:"result"`
	}
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	doc, err := parseWriteResult(ctx, resp.Result, "CmfDocument.update", c.documentByID, documentHasEmptyID)
	if err != nil {
		return nil, err
	}

	if hasText {
		if pubErr := c.DocumentPublish(ctx, doc.ID); pubErr != nil {
			return doc, errors.WithMessagef(pubErr, "document %s updated, publish failed; do not retry update", doc.ID)
		}
	}

	return doc, nil
}

// DocumentDelete deletes a document by ID
// Example:
//
//	err := client.DocumentDelete(ctx, "CmfDocument:uuid")
func (c *Client) DocumentDelete(
	ctx context.Context,
	docID string,
) error {
	if docID == "" {
		return errors.New("docID is required")
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.delete",
		CallID:  newCallID(),
		Args:    []any{docID},
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  any    `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// DocumentPageTree retrieves the document page tree hierarchy starting from the given node.
// Returns a flat list of documents with parent_id and tree_node_is_branch fields
// for building a tree structure.
// Example:
//
//	docs, err := client.DocumentPageTree(ctx, "CmfDocument:uuid")
func (c *Client) DocumentPageTree(
	ctx context.Context,
	nodeID string,
) ([]models.Document, error) {
	kwargs := map[string]any{
		"node_id": nodeID,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfDocument.macros_page_tree_get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.DocumentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil
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
