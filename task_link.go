package evateamclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultTaskLinkFields for task link queries.
var DefaultTaskLinkFields = []string{
	"id", "class_name", "source_id", "target_id", "link_type",
	"cmf_created_at", "comment",
}

// TaskLinks retrieves task relationships.
func (c *Client) TaskLinks(ctx context.Context, taskCode string, fields []string) ([]models.CmfTaskLink, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskFields
	}

	kwargs := map[string]any{
		"filter": []any{
			[]any{"source_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
			[]any{"target_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
		},
		"fields": fields,
	}

	return c.TaskLinksList(ctx, kwargs)
}

// TaskLinksOutgoing retrieves links where task is source only.
func (c *Client) TaskLinksOutgoing(ctx context.Context, taskCode string, fields []string) ([]models.CmfTaskLink, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskLinkFields
	}

	kwargs := map[string]any{
		"filter":   []any{"source_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TaskLinksList(ctx, kwargs)
}

// TaskLinksIncoming retrieves links where task is target only.
func (c *Client) TaskLinksIncoming(ctx context.Context, taskCode string, fields []string) ([]models.CmfTaskLink, *models.CmfMeta, error) {
	if len(fields) == 0 {
		fields = DefaultTaskLinkFields
	}

	kwargs := map[string]any{
		"filter":   []any{"target_id", "==", fmt.Sprintf("CmfTask:%s", taskCode)},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.TaskLinksList(ctx, kwargs)
}

// TaskLinksList retrieves task links with custom filters.
func (c *Client) TaskLinksList(ctx context.Context, kwargs map[string]any) ([]models.CmfTaskLink, *models.CmfMeta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfTaskLink.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CmfTaskLinkListResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/?m=CmfTaskLink.list", reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
