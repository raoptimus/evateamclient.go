package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	evateamclient "github.com/raoptimus/evateamclient"
)

// CommentTools provides MCP tool handlers for comment operations.
type CommentTools struct {
	client *evateamclient.Client
}

// NewCommentTools creates a new CommentTools instance.
func NewCommentTools(client *evateamclient.Client) *CommentTools {
	return &CommentTools{client: client}
}

// CommentListInput represents input for eva_comment_list tool.
type CommentListInput struct {
	QueryInput
	TaskID   string `json:"task_id,omitempty"`
	TaskCode string `json:"task_code,omitempty"`
	AuthorID string `json:"author_id,omitempty"`
}

// CommentList returns a list of comments.
func (c *CommentTools) CommentList(ctx context.Context, input CommentListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityComment, &input.QueryInput)
	if err != nil {
		return nil, WrapError("comment_list", err)
	}

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"task_id": input.TaskID})
	}
	if input.TaskCode != "" {
		qb = qb.Where(sq.Eq{"task_id": "Task:" + input.TaskCode})
	}
	if input.AuthorID != "" {
		qb = qb.Where(sq.Eq{"cmf_author_id": input.AuthorID})
	}

	comments, _, err := c.client.CommentsList(ctx, qb)
	if err != nil {
		return nil, WrapError("comment_list", err)
	}

	return &ListResult{
		Items:   comments,
		HasMore: len(comments) == input.Limit && input.Limit > 0,
	}, nil
}

// CommentGetInput represents input for eva_comment_get tool.
type CommentGetInput struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields,omitempty"`
}

// CommentGet retrieves a single comment.
func (c *CommentTools) CommentGet(ctx context.Context, input CommentGetInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("comment_get", ErrInvalidInput)
	}

	comment, _, err := c.client.Comment(ctx, input.ID, input.Fields)
	if err != nil {
		return nil, WrapError("comment_get", err)
	}

	return comment, nil
}

// CommentCreateInput represents input for eva_comment_create tool.
type CommentCreateInput struct {
	TaskID string `json:"task_id"`
	Text   string `json:"text"`
}

// CommentCreate creates a new comment.
func (c *CommentTools) CommentCreate(ctx context.Context, input CommentCreateInput) (any, error) {
	if input.TaskID == "" || input.Text == "" {
		return nil, WrapError("comment_create", ErrInvalidInput)
	}

	comment, err := c.client.CommentCreate(ctx, input.TaskID, input.Text)
	if err != nil {
		return nil, WrapError("comment_create", err)
	}

	return comment, nil
}

// CommentUpdateInput represents input for eva_comment_update tool.
type CommentUpdateInput struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// CommentUpdate updates an existing comment.
func (c *CommentTools) CommentUpdate(ctx context.Context, input CommentUpdateInput) (any, error) {
	if input.ID == "" || input.Text == "" {
		return nil, WrapError("comment_update", ErrInvalidInput)
	}

	comment, err := c.client.CommentUpdate(ctx, input.ID, input.Text)
	if err != nil {
		return nil, WrapError("comment_update", err)
	}

	return comment, nil
}

// CommentDeleteInput represents input for eva_comment_delete tool.
type CommentDeleteInput struct {
	ID string `json:"id"`
}

// CommentDelete deletes a comment.
func (c *CommentTools) CommentDelete(ctx context.Context, input CommentDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("comment_delete", ErrInvalidInput)
	}

	err := c.client.CommentDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("comment_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// CommentCountInput represents input for eva_comment_count tool.
type CommentCountInput struct {
	TaskID   string `json:"task_id,omitempty"`
	TaskCode string `json:"task_code,omitempty"`
	AuthorID string `json:"author_id,omitempty"`
}

// CommentCount counts comments.
func (c *CommentTools) CommentCount(ctx context.Context, input CommentCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityComment)

	if input.TaskID != "" {
		qb = qb.Where(sq.Eq{"task_id": input.TaskID})
	}
	if input.TaskCode != "" {
		qb = qb.Where(sq.Eq{"task_id": "Task:" + input.TaskCode})
	}
	if input.AuthorID != "" {
		qb = qb.Where(sq.Eq{"cmf_author_id": input.AuthorID})
	}

	count, err := c.client.CommentCount(ctx, qb)
	if err != nil {
		return nil, WrapError("comment_count", err)
	}

	return &CountResult{Count: count}, nil
}
