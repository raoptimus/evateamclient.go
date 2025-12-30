package evateamclient

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// QueryBuilder wraps Squirrel's SelectBuilder and converts to EVA API kwargs
// This allows using real Squirrel API with EVA Team backend
//
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "name", "code").
//	  From(EntityProject).
//	  Where(sq.Eq{"code": "PROJ-123"}).
//	  OrderBy("-cmf_created_at").
//	  Limit(50)
//
//	projects, meta, err := client.ProjectsList(ctx, qb)
type QueryBuilder struct {
	selectBuilder sq.SelectBuilder
	includeArch   bool
	noMeta        bool
}

// NewQueryBuilder creates a new EVA-compatible Squirrel builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		selectBuilder: sq.Select("*"),
	}
}

// Select sets columns to retrieve (maps to EVA "fields")
// Example: qb.Select("id", "name", "code", "executors")
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.Columns(columns...)
	return qb
}

// From sets the entity type (required for EVA API method routing)
// Valid values: EntityProject, EntityTask, EntityDocument, etc.
// Example: qb.From(EntityProject)
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.From(table)
	return qb
}

// Where adds filter conditions using Squirrel predicates
// Multiple Where() calls are combined with AND logic
//
// Examples:
//
//	qb.Where(sq.Eq{"code": "PROJ-123"})
//	qb.Where(sq.Gt{"priority": 3})
//	qb.Where(sq.Like{"name": "%Mobile%"})
//	qb.Where(sq.And{sq.Eq{"system": false}, sq.GtOrEq{"created_at": "2024-01-01"}})
func (qb *QueryBuilder) Where(pred any) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.Where(pred)
	return qb
}

// OrderBy adds sorting
// Use "-field" prefix for DESC order, "field" for ASC
//
// Examples:
//
//	qb.OrderBy("name")                    // ASC
//	qb.OrderBy("-cmf_created_at")        // DESC
//	qb.OrderBy("-priority", "name")      // Multiple columns
func (qb *QueryBuilder) OrderBy(orderBys ...string) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.OrderBy(orderBys...)
	return qb
}

// Limit sets maximum number of results
// Example: qb.Limit(50)
func (qb *QueryBuilder) Limit(limit uint64) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.Limit(limit)
	return qb
}

// Offset sets result offset for pagination
// Example: qb.Offset(100).Limit(50) // Skip 100, take 50
func (qb *QueryBuilder) Offset(offset uint64) *QueryBuilder {
	qb.selectBuilder = qb.selectBuilder.Offset(offset)
	return qb
}

// IncludeArchived includes deleted/archived objects (EVA-specific)
// Example: qb.Where(sq.Eq{"cmf_deleted": true}).IncludeArchived()
func (qb *QueryBuilder) IncludeArchived() *QueryBuilder {
	qb.includeArch = true
	return qb
}

// NoMeta disables meta response (EVA-specific, faster queries)
// Example: qb.NoMeta() // Skip metadata for better performance
func (qb *QueryBuilder) NoMeta() *QueryBuilder {
	qb.noMeta = true
	return qb
}

// ToKwargs converts Squirrel SelectBuilder to EVA API kwargs
// This translates SQL-like queries to JSON-RPC BQL format
//
// Returns map with keys: filter, fields, order_by, slice, include_archived, no_meta
func (qb *QueryBuilder) ToKwargs() (map[string]any, error) {
	kwargs := make(map[string]any)

	// Extract parts from Squirrel builder
	sqlStr, args, err := qb.selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	// Parse SQL to extract EVA components
	parts := parseSquirrelSQL(sqlStr, args)

	// Convert WHERE clause to EVA filter
	if len(parts.filters) > 0 {
		if len(parts.filters) == 1 {
			kwargs["filter"] = parts.filters[0]
		} else {
			kwargs["filter"] = parts.filters
		}
	}

	// Convert SELECT columns to EVA fields
	if len(parts.fields) > 0 && parts.fields[0] != "*" {
		kwargs["fields"] = parts.fields
	}

	// Convert ORDER BY to EVA order_by
	if len(parts.orderBy) > 0 {
		kwargs["order_by"] = parts.orderBy
	}

	// Convert LIMIT/OFFSET to EVA slice
	if parts.limit > 0 || parts.offset > 0 {
		kwargs["slice"] = []uint64{parts.offset, parts.limit}
	}

	// EVA-specific flags
	if qb.includeArch {
		kwargs["include_archived"] = true
	}
	if qb.noMeta {
		kwargs["no_meta"] = true
	}

	return kwargs, nil
}

// ToMethod returns the appropriate EVA API method based on table
// Example: "CmfProject" -> "CmfProject.list"
func (qb *QueryBuilder) ToMethod() (string, error) {
	// Extract table name from Squirrel builder
	sqlStr, _, err := qb.selectBuilder.ToSql()
	if err != nil {
		return "", err
	}

	table := extractTableName(sqlStr)
	if table == "" {
		return "", fmt.Errorf("table name not found in query, use From()")
	}

	// Determine method based on query type
	// If has LIMIT 1 or specific ID filter, use .get, otherwise .list
	return table + ".list", nil
}

// Validate checks if query is valid before execution
// Returns error if query has invalid parameters
func (qb *QueryBuilder) Validate() error {
	// Check if From() was called
	sqlStr, _, err := qb.selectBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("invalid query: %w", err)
	}

	if !strings.Contains(sqlStr, " FROM ") {
		return fmt.Errorf("missing From() clause - specify entity type")
	}

	return nil
}

// String returns human-readable query representation (for debugging)
func (qb *QueryBuilder) String() string {
	sqlStr, args, _ := qb.selectBuilder.ToSql()
	return fmt.Sprintf("QueryBuilder{sql=%q, args=%v, includeArch=%v, noMeta=%v}",
		sqlStr, args, qb.includeArch, qb.noMeta)
}

// sqlParts holds parsed SQL components
type sqlParts struct {
	fields  []string
	table   string
	filters []any
	orderBy []string
	limit   uint64
	offset  uint64
}

// parseSquirrelSQL converts Squirrel SQL to EVA BQL components
// This is a simplified parser - production version needs more robust parsing
func parseSquirrelSQL(sqlStr string, args []any) *sqlParts {
	parts := &sqlParts{
		fields:  []string{},
		filters: []any{},
		orderBy: []string{},
	}

	// Extract SELECT columns
	if idx := strings.Index(sqlStr, "SELECT "); idx >= 0 {
		fromIdx := strings.Index(sqlStr, " FROM ")
		if fromIdx > 0 {
			colsStr := strings.TrimSpace(sqlStr[idx+7 : fromIdx])
			if colsStr != "*" {
				parts.fields = strings.Split(colsStr, ", ")
			}
		}
	}

	// Extract table name
	parts.table = extractTableName(sqlStr)

	// Extract WHERE conditions
	if whereIdx := strings.Index(sqlStr, "WHERE "); whereIdx >= 0 {
		parts.filters = convertSquirrelFilters(sqlStr[whereIdx:], args)
	}

	// Extract ORDER BY
	if orderIdx := strings.Index(sqlStr, "ORDER BY "); orderIdx >= 0 {
		limitIdx := strings.Index(sqlStr[orderIdx:], " LIMIT ")
		endIdx := len(sqlStr)
		if limitIdx > 0 {
			endIdx = orderIdx + limitIdx
		}
		orderStr := strings.TrimSpace(sqlStr[orderIdx+9 : endIdx])
		parts.orderBy = parseOrderBy(orderStr)
	}

	// Extract LIMIT/OFFSET
	if limitIdx := strings.Index(sqlStr, "LIMIT "); limitIdx >= 0 {
		fmt.Sscanf(sqlStr[limitIdx:], "LIMIT %d", &parts.limit)
	}
	if offsetIdx := strings.Index(sqlStr, "OFFSET "); offsetIdx >= 0 {
		fmt.Sscanf(sqlStr[offsetIdx:], "OFFSET %d", &parts.offset)
	}

	return parts
}

// extractTableName extracts table name from SQL string
func extractTableName(sqlStr string) string {
	fromIdx := strings.Index(sqlStr, " FROM ")
	if fromIdx < 0 {
		return ""
	}

	afterFrom := sqlStr[fromIdx+6:]
	whereIdx := strings.Index(afterFrom, " WHERE ")
	orderIdx := strings.Index(afterFrom, " ORDER BY ")
	limitIdx := strings.Index(afterFrom, " LIMIT ")

	endIdx := len(afterFrom)
	if whereIdx > 0 && whereIdx < endIdx {
		endIdx = whereIdx
	}
	if orderIdx > 0 && orderIdx < endIdx {
		endIdx = orderIdx
	}
	if limitIdx > 0 && limitIdx < endIdx {
		endIdx = limitIdx
	}

	return strings.TrimSpace(afterFrom[:endIdx])
}

// convertSquirrelFilters converts SQL WHERE to EVA BQL filters
// Handles Squirrel's Eq, Gt, Lt, Like, etc.
func convertSquirrelFilters(whereClause string, args []any) []any {
	filters := []any{}

	// Simple parser for common patterns
	argIdx := 0

	// Pattern: field = ?
	if strings.Contains(whereClause, " = ?") {
		parts := strings.Split(whereClause, " = ?")
		for i, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, "==", args[argIdx]})
				argIdx++
			}

			// Check for AND
			if i < len(parts)-2 && strings.Contains(parts[i+1], " AND ") {
				// Multiple filters, continue
			}
		}
	}

	// Pattern: field > ?
	if strings.Contains(whereClause, " > ?") {
		parts := strings.Split(whereClause, " > ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, ">", args[argIdx]})
				argIdx++
			}
		}
	}

	// Pattern: field >= ?
	if strings.Contains(whereClause, " >= ?") {
		parts := strings.Split(whereClause, " >= ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, ">=", args[argIdx]})
				argIdx++
			}
		}
	}

	// Pattern: field < ?
	if strings.Contains(whereClause, " < ?") {
		parts := strings.Split(whereClause, " < ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, "<", args[argIdx]})
				argIdx++
			}
		}
	}

	// Pattern: field <= ?
	if strings.Contains(whereClause, " <= ?") {
		parts := strings.Split(whereClause, " <= ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, "<=", args[argIdx]})
				argIdx++
			}
		}
	}

	// Pattern: field != ?
	if strings.Contains(whereClause, " != ?") {
		parts := strings.Split(whereClause, " != ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, "!=", args[argIdx]})
				argIdx++
			}
		}
	}

	// Pattern: field LIKE ?
	if strings.Contains(whereClause, " LIKE ?") {
		parts := strings.Split(whereClause, " LIKE ?")
		for _, part := range parts[:len(parts)-1] {
			fieldName := extractLastWord(part)
			if argIdx < len(args) {
				filters = append(filters, []any{fieldName, "LIKE", args[argIdx]})
				argIdx++
			}
		}
	}

	return filters
}

// extractLastWord extracts the last word from a string (field name)
func extractLastWord(s string) string {
	s = strings.TrimSpace(s)
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	return words[len(words)-1]
}

// parseOrderBy converts SQL ORDER BY to EVA format
// SQL: "created_at DESC, name ASC" -> EVA: ["-created_at", "name"]
func parseOrderBy(orderStr string) []string {
	result := []string{}
	parts := strings.Split(orderStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasSuffix(part, " DESC") {
			field := strings.TrimSuffix(part, " DESC")
			result = append(result, "-"+strings.TrimSpace(field))
		} else if strings.HasSuffix(part, " ASC") {
			field := strings.TrimSuffix(part, " ASC")
			result = append(result, strings.TrimSpace(field))
		} else {
			result = append(result, part)
		}
	}

	return result
}

// Helper functions for common Squirrel patterns with EVA compatibility

// Between creates a range filter for EVA using Squirrel's And combinator
// Example: qb.Where(Between("cmf_created_at", "2024-01-01", "2024-12-31"))
func Between(col string, from, to any) sq.And {
	return sq.And{
		sq.GtOrEq{col: from},
		sq.LtOrEq{col: to},
	}
}
