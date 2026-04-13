package models

import "time"

// Comment represents COMPLETE comment object.
type Comment struct {
	ID        string    `json:"id"`
	ClassName string    `json:"class_name"`
	AuthorID  string    `json:"cmf_author_id,omitempty"` // example: CmfPerson:36253940-a34a-11f0-9dac-5269014ed76a
	CreatedAt time.Time `json:"cmf_created_at,omitempty"`
	LogLevel  int       `json:"log_level,omitempty"`
	Text      string    `json:"text,omitempty"`
	Private   bool      `json:"private,omitempty"`
	ParentID  string    `json:"parent_id,omitempty"`    // example: CmfTask:06506d44-c545-11f0-b6f8-eeb7fce6ef9e
	ProjectID string    `json:"project_id,omitempty"`   // example: CmfProject:06506d44-c545-11f0-b6f8-eeb7fce6ef9e
	OwnerID   string    `json:"cmf_owner_id,omitempty"` // example: CmfPerson:06506d44-c545-11f0-b6f8-eeb7fce6ef9e
	Name      string    `json:"name,omitempty"`
	Code      string    `json:"code,omitempty"`
}

// CommentResponse for CmfComment.get (single comment).
type CommentResponse struct {
	JSONRPC string  `json:"jsonrpc,omitempty"`
	Result  Comment `json:"result,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`
}

// CommentListResponse for Comment.list.
type CommentListResponse struct {
	JSONRPC string    `json:"jsonrpc,omitempty"`
	Result  []Comment `json:"result,omitempty"`
	Meta    Meta      `json:"meta,omitempty"`
}
