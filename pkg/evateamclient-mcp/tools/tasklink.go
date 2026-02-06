package tools

import (
	"context"

	"github.com/raoptimus/evateamclient.go"
)

// TaskLinkTools provides MCP tool handlers for task link operations.
type TaskLinkTools struct {
	client *evateamclient.Client
}

// NewTaskLinkTools creates a new TaskLinkTools instance.
func NewTaskLinkTools(client *evateamclient.Client) *TaskLinkTools {
	return &TaskLinkTools{client: client}
}

// TaskLinkListInput represents input for eva_tasklink_list tool.
type TaskLinkListInput struct {
	QueryInput
	TaskID    string `json:"task_id,omitempty"`
	Direction string `json:"direction,omitempty"` // "outgoing", "incoming", "both" (default)
}

// TaskLinkList returns a list of task links.
func (t *TaskLinkTools) TaskLinkList(ctx context.Context, input *TaskLinkListInput) (*ListResult, error) {
	if input.TaskID == "" {
		// List all links with custom query
		qb, err := BuildQuery(evateamclient.EntityRelation, &input.QueryInput)
		if err != nil {
			return nil, WrapError("tasklink_list", err)
		}

		links, _, err := t.client.TaskLinksListQuery(ctx, qb)
		if err != nil {
			return nil, WrapError("tasklink_list", err)
		}

		return &ListResult{
			Items:   toAnySlice(links),
			HasMore: len(links) == input.Limit && input.Limit > 0,
		}, nil
	}

	// List links for specific task
	var (
		links []any
		err   error
	)

	switch input.Direction {
	case "outgoing":
		result, _, e := t.client.TaskLinksOutgoing(ctx, input.TaskID, nil)
		links, err = toAnySlice(result), e
	case "incoming":
		result, _, e := t.client.TaskLinksIncoming(ctx, input.TaskID, nil)
		links, err = toAnySlice(result), e
	default: // "both" or empty
		result, _, e := t.client.TaskLinks(ctx, input.TaskID, nil)
		links, err = toAnySlice(result), e
	}

	if err != nil {
		return nil, WrapError("tasklink_list", err)
	}

	return &ListResult{
		Items:   links,
		HasMore: false,
	}, nil
}

// TaskLinkGetInput represents input for eva_tasklink_get tool.
type TaskLinkGetInput struct {
	ID     string   `json:"id"`
	Fields []string `json:"fields,omitempty"`
}

// TaskLinkGet retrieves a single task link.
func (t *TaskLinkTools) TaskLinkGet(ctx context.Context, input TaskLinkGetInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("tasklink_get", ErrInvalidInput)
	}

	link, _, err := t.client.TaskLink(ctx, input.ID, input.Fields)
	if err != nil {
		return nil, WrapError("tasklink_get", err)
	}

	return link, nil
}

// TaskLinkCreateInput represents input for eva_tasklink_create tool.
type TaskLinkCreateInput struct {
	SourceTaskID     string `json:"source_task_id"`
	TargetTaskID     string `json:"target_task_id"`
	RelationOptionID string `json:"relation_option_id"`
}

// TaskLinkCreate creates a new task link.
func (t *TaskLinkTools) TaskLinkCreate(ctx context.Context, input TaskLinkCreateInput) (any, error) {
	if input.SourceTaskID == "" || input.TargetTaskID == "" || input.RelationOptionID == "" {
		return nil, WrapError("tasklink_create", ErrInvalidInput)
	}

	link, err := t.client.TaskLinkCreate(ctx, input.SourceTaskID, input.TargetTaskID, input.RelationOptionID)
	if err != nil {
		return nil, WrapError("tasklink_create", err)
	}

	return link, nil
}

// TaskLinkDeleteInput represents input for eva_tasklink_delete tool.
type TaskLinkDeleteInput struct {
	ID string `json:"id"`
}

// TaskLinkDelete deletes a task link.
func (t *TaskLinkTools) TaskLinkDelete(ctx context.Context, input TaskLinkDeleteInput) (any, error) {
	if input.ID == "" {
		return nil, WrapError("tasklink_delete", ErrInvalidInput)
	}

	err := t.client.TaskLinkDelete(ctx, input.ID)
	if err != nil {
		return nil, WrapError("tasklink_delete", err)
	}

	return map[string]bool{"success": true}, nil
}

// TaskLinkCountInput represents input for eva_tasklink_count tool.
type TaskLinkCountInput struct {
	TaskID    string `json:"task_id,omitempty"`
	Direction string `json:"direction,omitempty"` // "outgoing", "incoming", "both"
}

// TaskLinkCount counts task links.
func (t *TaskLinkTools) TaskLinkCount(ctx context.Context, input TaskLinkCountInput) (*CountResult, error) {
	if input.TaskID == "" {
		qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityRelation)
		count, err := t.client.TaskLinkCount(ctx, qb)
		if err != nil {
			return nil, WrapError("tasklink_count", err)
		}
		return &CountResult{Count: count}, nil
	}

	// Count links for specific task - need to query and count
	links, _, err := t.client.TaskLinks(ctx, input.TaskID, []string{"id"})
	if err != nil {
		return nil, WrapError("tasklink_count", err)
	}

	return &CountResult{Count: len(links)}, nil
}
