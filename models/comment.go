package models

// Comment represents COMPLETE comment object.
type Comment struct {
	ID        string `json:"id"`
	ClassName string `json:"class_name"`
	TaskID    string `json:"task_id,omitempty"`
	Text      string `json:"text,omitempty"`
	AuthorID  string `json:"cmf_author_id,omitempty"`
	CreatedAt string `json:"cmf_created_at,omitempty"`
	LogLevel  string `json:"log_level,omitempty"`
}

type CommentListResponse struct {
	JSONRPC string    `json:"jsonrpc,omitempty"`
	Result  []Comment `json:"result,omitempty"`
	Meta    Meta      `json:"meta,omitempty"`
}
