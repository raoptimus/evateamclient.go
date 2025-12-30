package evateamclient

import (
	"context"
	"fmt"

	"github.com/raoptimus/evateamclient/models"
)

// DefaultCommentFields for comment queries.
var DefaultCommentFields = []string{"id", "text", "cmf_author_id", "cmf_created_at"}

// TaskComments retrieves ALL comments for task.
func (c *Client) TaskComments(
	ctx context.Context,
	taskCode string,
	fields []string,
) ([]models.Comment, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultCommentFields
	}

	kwargs := map[string]any{
		"filter":   []any{"task_id", "==", fmt.Sprintf("Task:%s", taskCode)},
		"fields":   fields,
		"order_by": []string{"-cmf_created_at"},
	}

	return c.Comments(ctx, kwargs)
}

// Comments retrieves comments with custom filters.
func (c *Client) Comments(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Comment, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "Comment.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.CommentListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
