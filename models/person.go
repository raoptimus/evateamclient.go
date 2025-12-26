package models

// CmfPerson represents COMPLETE person object.
type CmfPerson struct {
	ID          string `json:"id"`         // "CmfPerson:UUID"
	ClassName   string `json:"class_name"` // "CmfPerson"
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	Login       string `json:"login,omitempty"`
	Active      bool   `json:"active,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Position    string `json:"position,omitempty"`
	Department  string `json:"department,omitempty"`
	ManagerID   string `json:"manager_id,omitempty"`
	HireDate    string `json:"hire_date,omitempty"`
	CacheStatus string `json:"cache_status_type,omitempty"`
}

// CmfPersonListResponse for CmfPerson.list.
type CmfPersonListResponse struct {
	JSONRPC string      `json:"jsonrpc,omitempty"`
	Result  []CmfPerson `json:"result,omitempty"`
	Meta    CmfMeta     `json:"meta,omitempty"`
}
