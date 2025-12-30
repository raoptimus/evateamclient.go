package models

// TaskLink represents COMPLETE task relationship.
type TaskLink struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id,omitempty"`
	TargetID string `json:"target_id,omitempty"`
	LinkType string `json:"link_type,omitempty"`
}

// TaskLinkListResponse for TaskLink.list.
type TaskLinkListResponse struct {
	JSONRPC string     `json:"jsonrpc,omitempty"`
	Result  []TaskLink `json:"result,omitempty"`
	Meta    Meta       `json:"meta,omitempty"`
}
