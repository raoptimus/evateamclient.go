// Package tools provides MCP tool handlers for EVA Team API.
package tools

// Filter represents a single filter condition for queries.
// Format: [field, operator, value]
// Operators: "==", "!=", ">", ">=", "<", "<=", "LIKE", "contains"
type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// QueryInput represents common query parameters for list operations.
type QueryInput struct {
	// Fields to return (projection)
	Fields []string `json:"fields,omitempty"`

	// Filters for query
	Filters []Filter `json:"filters,omitempty"`

	// Sorting: use "-field" for DESC, "field" for ASC
	OrderBy []string `json:"order_by,omitempty"`

	// Pagination
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`

	// Include archived/deleted records
	IncludeArchived bool `json:"include_archived,omitempty"`
}

// ListResult wraps list response with metadata.
type ListResult struct {
	Items      any   `json:"items"`
	TotalCount int64 `json:"total_count,omitempty"`
	HasMore    bool  `json:"has_more,omitempty"`
}

// CountResult wraps count response.
type CountResult struct {
	Count int `json:"count"`
}
