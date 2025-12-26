package models

// CmfTaskLink represents COMPLETE task relationship.
type CmfTaskLink struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id,omitempty"`
	TargetID string `json:"target_id,omitempty"`
	LinkType string `json:"link_type,omitempty"`
}

// CmfTaskLinkListResponse for CmfTaskLink.list.
type CmfTaskLinkListResponse struct {
	JSONRPC string        `json:"jsonrpc,omitempty"`
	Result  []CmfTaskLink `json:"result,omitempty"`
	Meta    CmfMeta       `json:"meta,omitempty"`
}
