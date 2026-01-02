package evateamclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Internal function tests for query_builder.go

func TestExtractTableName_SimpleSelect_ReturnsTable(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{"simple select", "SELECT * FROM CmfTask", "CmfTask"},
		{"with where", "SELECT * FROM CmfProject WHERE code = ?", "CmfProject"},
		{"with order", "SELECT * FROM CmfTask ORDER BY name", "CmfTask"},
		{"with limit", "SELECT * FROM CmfTask LIMIT 10", "CmfTask"},
		{"with all clauses", "SELECT id FROM CmfTask WHERE id = ? ORDER BY name LIMIT 10", "CmfTask"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTableName(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTableName_NoFrom_ReturnsEmpty(t *testing.T) {
	result := extractTableName("SELECT * WHERE code = ?")
	assert.Empty(t, result)
}

func TestParseOrderBy_SQLFormat_ReturnsEVAFormat(t *testing.T) {
	tests := []struct {
		name     string
		orderStr string
		expected []string
	}{
		{"ascending explicit", "name ASC", []string{"name"}},
		{"descending explicit", "created_at DESC", []string{"-created_at"}},
		{"ascending implicit", "name", []string{"name"}},
		{"multiple columns", "priority DESC, name ASC", []string{"-priority", "name"}},
		{"multiple with implicit", "priority DESC, name", []string{"-priority", "name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOrderBy(tt.orderStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractLastWord_ValidStrings_ReturnsLastWord(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"field", "field"},
		{"WHERE field", "field"},
		{"AND field2", "field2"},
		{"  spaced  ", "spaced"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractLastWord(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertSquirrelFilters_EqualityOperator_ReturnsFilter(t *testing.T) {
	whereClause := "WHERE field = ?"
	args := []any{"value"}

	filters := convertSquirrelFilters(whereClause, args)

	assert.Len(t, filters, 1)
	filter := filters[0].([]any)
	assert.Equal(t, "field", filter[0])
	assert.Equal(t, "==", filter[1])
	assert.Equal(t, "value", filter[2])
}

func TestConvertSquirrelFilters_ComparisonOperators_ReturnsFilters(t *testing.T) {
	tests := []struct {
		name         string
		whereClause  string
		args         []any
		expectedOp   string
		expectedVal  any
	}{
		{"greater than", "WHERE priority > ?", []any{3}, ">", 3},
		{"greater or equal", "WHERE priority >= ?", []any{3}, ">=", 3},
		{"less than", "WHERE priority < ?", []any{5}, "<", 5},
		{"less or equal", "WHERE priority <= ?", []any{5}, "<=", 5},
		{"not equal", "WHERE status != ?", []any{"CLOSED"}, "!=", "CLOSED"},
		{"like", "WHERE name LIKE ?", []any{"%test%"}, "LIKE", "%test%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := convertSquirrelFilters(tt.whereClause, tt.args)

			assert.NotEmpty(t, filters, "filter should not be empty for %s", tt.name)
			if len(filters) > 0 {
				filter := filters[0].([]any)
				assert.Equal(t, tt.expectedOp, filter[1])
				assert.Equal(t, tt.expectedVal, filter[2])
			}
		})
	}
}

func TestParseSquirrelSQL_ExtractsFields(t *testing.T) {
	sqlStr := "SELECT id, name, code FROM CmfTask"
	args := []any{}

	parts := parseSquirrelSQL(sqlStr, args)

	assert.Equal(t, []string{"id", "name", "code"}, parts.fields)
	assert.Equal(t, "CmfTask", parts.table)
}

func TestParseSquirrelSQL_WildcardFields_ReturnsEmpty(t *testing.T) {
	sqlStr := "SELECT * FROM CmfTask"
	args := []any{}

	parts := parseSquirrelSQL(sqlStr, args)

	assert.Empty(t, parts.fields)
	assert.Equal(t, "CmfTask", parts.table)
}

func TestParseSquirrelSQL_ExtractsLimitOffset(t *testing.T) {
	sqlStr := "SELECT * FROM CmfTask LIMIT 50 OFFSET 10"
	args := []any{}

	parts := parseSquirrelSQL(sqlStr, args)

	assert.Equal(t, uint64(50), parts.limit)
	assert.Equal(t, uint64(10), parts.offset)
}

func TestParseSquirrelSQL_ExtractsOrderBy(t *testing.T) {
	sqlStr := "SELECT * FROM CmfTask ORDER BY priority DESC, name ASC"
	args := []any{}

	parts := parseSquirrelSQL(sqlStr, args)

	assert.Equal(t, []string{"-priority", "name"}, parts.orderBy)
}

func TestParseSquirrelSQL_CompleteQuery(t *testing.T) {
	sqlStr := "SELECT id, name FROM CmfTask WHERE status = ? ORDER BY name DESC LIMIT 100 OFFSET 50"
	args := []any{"OPEN"}

	parts := parseSquirrelSQL(sqlStr, args)

	assert.Equal(t, []string{"id", "name"}, parts.fields)
	assert.Equal(t, "CmfTask", parts.table)
	assert.NotEmpty(t, parts.filters)
	assert.Equal(t, []string{"-name"}, parts.orderBy)
	assert.Equal(t, uint64(100), parts.limit)
	assert.Equal(t, uint64(50), parts.offset)
}
