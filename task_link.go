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

// RelationTypeLink is the well-known "Относится к" (Relates to) task-link
// relation_type code, per KB-000323 (doc/primery_api_zaprosov_doc-000244.pdf).
// No other codes are documented; do not invent additional ones.
const RelationTypeLink = "system.link"

// TaskLink field constants for type-safe queries
const (
	// Core fields
	TaskLinkFieldID           = "id"
	TaskLinkFieldClassName    = "class_name"
	TaskLinkFieldCode         = "code"
	TaskLinkFieldName         = "name"
	TaskLinkFieldInLink       = "in_link"       // filter field: incoming links to task
	TaskLinkFieldOutLink      = "out_link"      // filter field: outgoing links from task
	TaskLinkFieldRelationType = "relation_type" // link type code, e.g. RelationTypeLink

	// System
	TaskLinkFieldCmfCreatedAt  = "cmf_created_at"
	TaskLinkFieldCmfModifiedAt = "cmf_modified_at"
	TaskLinkFieldCmfOwnerID    = "cmf_owner_id"
)

var (
	// DefaultTaskLinkFields - standard projection for task link queries
	DefaultTaskLinkFields = []string{
		TaskLinkFieldID,
		TaskLinkFieldClassName,
		TaskLinkFieldCode,
		TaskLinkFieldName,
		TaskLinkFieldInLink,
		TaskLinkFieldOutLink,
		TaskLinkFieldRelationType,
		TaskLinkFieldCmfCreatedAt,
		TaskLinkFieldCmfOwnerID,
	}

	// DefaultTaskLinkListFields - optimized for LIST queries
	DefaultTaskLinkListFields = []string{
		TaskLinkFieldID,
		TaskLinkFieldCode,
		TaskLinkFieldName,
		TaskLinkFieldInLink,
		TaskLinkFieldOutLink,
		TaskLinkFieldRelationType,
	}
)

// TaskLink retrieves a single task link by ID
// Example:
//
//	link, meta, err := client.TaskLink(ctx, "CmfTaskLink:uuid", nil)
func (c *Client) TaskLink(
	ctx context.Context,
	linkID string,
	fields []string,
) (*models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldID: linkID}).
		Limit(1)

	return c.TaskLinkQuery(ctx, qb)
}

// TaskLinkQuery executes query using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{"id": "CmfRelationOption:uuid"})
//	link, meta, err := client.TaskLinkQuery(ctx, qb)
func (c *Client) TaskLinkQuery(ctx context.Context, qb *QueryBuilder) (*models.TaskLink, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// TaskLinksListQuery retrieves list using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "code", "name").
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{evateamclient.TaskLinkFieldOutLink: "CmfTask:uuid"}).
//	  Limit(100)
//	links, meta, err := client.TaskLinksListQuery(ctx, qb)
func (c *Client) TaskLinksListQuery(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.TaskLink, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkListFields
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

	var resp models.TaskLinkListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// TaskLinkCount counts task links using QueryBuilder
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityRelation).
//	  Where(sq.Eq{evateamclient.TaskLinkFieldOutLink: "CmfTask:uuid"})
//	count, err := client.TaskLinkCount(ctx, qb)
func (c *Client) TaskLinkCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.count",
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

// TaskLinks retrieves ALL task relationships (both directions)
// Makes two API calls (outgoing + incoming) and merges results
// Example:
//
//	links, meta, err := client.TaskLinks(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinks(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	// Get outgoing links (task is source)
	outgoing, _, err := c.TaskLinksOutgoing(ctx, taskID, fields)
	if err != nil {
		return nil, nil, err
	}

	// Get incoming links (task is target)
	incoming, meta, err := c.TaskLinksIncoming(ctx, taskID, fields)
	if err != nil {
		return nil, nil, err
	}

	// Merge results, avoiding duplicates by ID
	seen := make(map[string]bool)
	var result []models.TaskLink

	for i := range outgoing {
		if !seen[outgoing[i].ID] {
			seen[outgoing[i].ID] = true
			result = append(result, outgoing[i])
		}
	}
	for i := range incoming {
		if !seen[incoming[i].ID] {
			seen[incoming[i].ID] = true
			result = append(result, incoming[i])
		}
	}

	return result, meta, nil
}

// TaskLinksOutgoing retrieves links where task is source (outgoing)
// Example:
//
//	links, meta, err := client.TaskLinksOutgoing(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinksOutgoing(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldOutLink: taskID})

	return c.TaskLinksListQuery(ctx, qb)
}

// TaskLinksIncoming retrieves links where task is target (incoming)
// Example:
//
//	links, meta, err := client.TaskLinksIncoming(ctx, "CmfTask:uuid", nil)
func (c *Client) TaskLinksIncoming(
	ctx context.Context,
	taskID string,
	fields []string,
) ([]models.TaskLink, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityRelation).
		Where(sq.Eq{TaskLinkFieldInLink: taskID})

	return c.TaskLinksListQuery(ctx, qb)
}

// TaskLinkCreate creates a link between two tasks.
//
// outLink is the source task, inLink is the target task; both accept a task
// code (e.g. "TSK-000001") or ID (e.g. "CmfTask:uuid"). relationType is the
// link type code (e.g. RelationTypeLink), not a relation-option ID/code.
//
// Example:
//
//	link, err := client.TaskLinkCreate(ctx, "TSK-000001", "TSK-000002", evateamclient.RelationTypeLink)
func (c *Client) TaskLinkCreate(
	ctx context.Context,
	outLink, inLink, relationType string,
) (*models.TaskLink, error) {
	if outLink == "" {
		return nil, errors.New("outLink is required")
	}
	if inLink == "" {
		return nil, errors.New("inLink is required")
	}
	if relationType == "" {
		return nil, errors.New("relationType is required")
	}

	kwargs := map[string]any{
		TaskLinkFieldOutLink:      outLink,
		TaskLinkFieldInLink:       inLink,
		TaskLinkFieldRelationType: relationType,
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.create",
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

	return parseWriteResult(ctx, resp.Result, "CmfRelationOption.create", c.fetchTaskLinkByIDOrCode, taskLinkHasEmptyID)
}

// fetchTaskLinkByIDOrCode resolves the bare string returned by
// CmfRelationOption.create — an ID ("CmfRelationOption:uuid") or a code
// ("RLO-000123"), distinguished by the ":" class-name prefix only IDs carry.
func (c *Client) fetchTaskLinkByIDOrCode(ctx context.Context, idOrCode string) (*models.TaskLink, error) {
	field := TaskLinkFieldCode
	if strings.Contains(idOrCode, ":") {
		field = TaskLinkFieldID
	}

	qb := NewQueryBuilder().
		Select(DefaultTaskLinkFields...).
		From(EntityRelation).
		Where(sq.Eq{field: idOrCode}).
		Limit(1)

	link, _, err := c.TaskLinkQuery(ctx, qb)
	return link, err
}

func taskLinkHasEmptyID(link *models.TaskLink) bool {
	return link == nil || link.ID == ""
}

// TaskLinkDelete deletes a task link by ID
// Example:
//
//	err := client.TaskLinkDelete(ctx, "RLO-000001")
func (c *Client) TaskLinkDelete(
	ctx context.Context,
	linkID string,
) error {
	if linkID == "" {
		return errors.New("linkID is required")
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.delete",
		CallID:  newCallID(),
		Args:    []any{linkID},
	}

	// CmfRelationOption.delete is undocumented in the OAS; the result shape is
	// unknown, so decode leniently like TaskDelete does.
	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  any    `json:"result"`
	}

	return c.doRequest(ctx, reqBody, &resp)
}

// Backward compatible methods (using old API)

// TaskLinksList retrieves task links with custom filters (backward compatible, deprecated)
// Recommended: use TaskLinksListQuery with NewQueryBuilder() instead
func (c *Client) TaskLinksList(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.TaskLink, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultTaskLinkListFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfRelationOption.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.TaskLinkListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
