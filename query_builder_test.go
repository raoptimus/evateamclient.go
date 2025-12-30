package evateamclient

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder_Select_SetsColumns_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		columns        []string
		expectedFields []string
	}{
		{
			name:           "single column",
			columns:        []string{"id"},
			expectedFields: []string{"*", "id"},
		},
		{
			name:           "multiple columns",
			columns:        []string{"id", "name", "code"},
			expectedFields: []string{"*", "id", "name", "code"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().
				Select(tt.columns...).
				From(EntityTask)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err)

			fields, ok := kwargs["fields"].([]string)
			require.True(t, ok)
			assert.Equal(t, tt.expectedFields, fields)
		})
	}
}

func TestQueryBuilder_From_SetsEntityType_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		entity         string
		expectedMethod string
	}{
		{
			name:           "project entity",
			entity:         EntityProject,
			expectedMethod: "CmfProject.list",
		},
		{
			name:           "task entity",
			entity:         EntityTask,
			expectedMethod: "CmfTask.list",
		},
		{
			name:           "time log entity",
			entity:         EntityTimeLog,
			expectedMethod: "CmfTimeTrackerHistory.list",
		},
		{
			name:           "status history entity",
			entity:         EntityStatusHistory,
			expectedMethod: "CmfStatusHistory.list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().From(tt.entity)

			method, err := qb.ToMethod()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestQueryBuilder_Where_AddsFilter_Successfully(t *testing.T) {
	tests := []struct {
		name           string
		predicate      any
		args           []any
		expectedFilter []any
	}{
		{
			name:           "equality filter",
			predicate:      sq.Eq{"code": "TASK-123"},
			expectedFilter: []any{"code", "==", "TASK-123"},
		},
		{
			name:           "greater than filter",
			predicate:      sq.Gt{"priority": 3},
			expectedFilter: []any{"priority", ">", 3},
		},
		{
			name:           "less than filter",
			predicate:      sq.Lt{"priority": 5},
			expectedFilter: []any{"priority", "<", 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().
				From(EntityTask).
				Where(tt.predicate)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err)

			filter, ok := kwargs["filter"]
			require.True(t, ok)
			assert.Equal(t, tt.expectedFilter, filter)
		})
	}
}

func TestQueryBuilder_OrderBy_SetsSorting_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		orderBy         []string
		expectedOrderBy []string
	}{
		{
			name:            "ascending order",
			orderBy:         []string{"name"},
			expectedOrderBy: []string{"name"},
		},
		{
			name:            "descending order with prefix",
			orderBy:         []string{"-cmf_created_at"},
			expectedOrderBy: []string{"-cmf_created_at"},
		},
		{
			name:            "multiple sort columns",
			orderBy:         []string{"-priority", "name"},
			expectedOrderBy: []string{"-priority", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().
				From(EntityTask).
				OrderBy(tt.orderBy...)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err)

			orderBy, ok := kwargs["order_by"].([]string)
			require.True(t, ok)
			assert.Equal(t, tt.expectedOrderBy, orderBy)
		})
	}
}

func TestQueryBuilder_LimitOffset_SetsSlice_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		limit         uint64
		offset        uint64
		expectedSlice []uint64
	}{
		{
			name:          "limit only",
			limit:         50,
			offset:        0,
			expectedSlice: []uint64{0, 50},
		},
		{
			name:          "offset and limit",
			limit:         100,
			offset:        50,
			expectedSlice: []uint64{50, 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder().
				From(EntityTask).
				Offset(tt.offset).
				Limit(tt.limit)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err)

			slice, ok := kwargs["slice"].([]uint64)
			require.True(t, ok)
			assert.Equal(t, tt.expectedSlice, slice)
		})
	}
}

func TestQueryBuilder_IncludeArchived_SetsFlag_Successfully(t *testing.T) {
	qb := NewQueryBuilder().
		From(EntityTask).
		IncludeArchived()

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	includeArch, ok := kwargs["include_archived"].(bool)
	require.True(t, ok)
	assert.True(t, includeArch)
}

func TestQueryBuilder_NoMeta_SetsFlag_Successfully(t *testing.T) {
	qb := NewQueryBuilder().
		From(EntityTask).
		NoMeta()

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	noMeta, ok := kwargs["no_meta"].(bool)
	require.True(t, ok)
	assert.True(t, noMeta)
}

func TestQueryBuilder_Validate_MissingFrom_ReturnsError(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "name")

	err := qb.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing From()")
}

func TestQueryBuilder_Validate_ValidQuery_ReturnsNoError(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "name").
		From(EntityTask)

	err := qb.Validate()
	assert.NoError(t, err)
}

func TestQueryBuilder_ToMethod_NoFrom_ReturnsError(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "name")

	_, err := qb.ToMethod()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name not found")
}

func TestQueryBuilder_String_ReturnsRepresentation_Successfully(t *testing.T) {
	qb := NewQueryBuilder().
		Select("id", "name").
		From(EntityTask).
		Where(sq.Eq{"code": "TEST"}).
		IncludeArchived()

	str := qb.String()
	assert.Contains(t, str, "QueryBuilder")
	assert.Contains(t, str, "includeArch=true")
}

func TestBetween_CreatesRangeFilter_Successfully(t *testing.T) {
	filter := Between("cmf_created_at", "2024-01-01", "2024-12-31")

	assert.Len(t, filter, 2)
	assert.IsType(t, sq.GtOrEq{}, filter[0])
	assert.IsType(t, sq.LtOrEq{}, filter[1])
}

func TestExtractTableName_ValidSQL_ReturnsTableName(t *testing.T) {
	tests := []struct {
		name          string
		sql           string
		expectedTable string
	}{
		{
			name:          "simple select",
			sql:           "SELECT * FROM CmfTask",
			expectedTable: "CmfTask",
		},
		{
			name:          "select with where",
			sql:           "SELECT * FROM CmfProject WHERE code = ?",
			expectedTable: "CmfProject",
		},
		{
			name:          "select with order",
			sql:           "SELECT * FROM CmfTask ORDER BY name",
			expectedTable: "CmfTask",
		},
		{
			name:          "select with limit",
			sql:           "SELECT * FROM CmfTask LIMIT 10",
			expectedTable: "CmfTask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := extractTableName(tt.sql)
			assert.Equal(t, tt.expectedTable, table)
		})
	}
}

func TestExtractTableName_NoFrom_ReturnsEmpty(t *testing.T) {
	table := extractTableName("SELECT * WHERE code = ?")
	assert.Empty(t, table)
}

func TestParseOrderBy_SQLFormat_ReturnsEVAFormat(t *testing.T) {
	tests := []struct {
		name           string
		orderStr       string
		expectedResult []string
	}{
		{
			name:           "ascending",
			orderStr:       "name ASC",
			expectedResult: []string{"name"},
		},
		{
			name:           "descending",
			orderStr:       "created_at DESC",
			expectedResult: []string{"-created_at"},
		},
		{
			name:           "multiple columns",
			orderStr:       "priority DESC, name ASC",
			expectedResult: []string{"-priority", "name"},
		},
		{
			name:           "no direction",
			orderStr:       "name",
			expectedResult: []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseOrderBy(tt.orderStr)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestExtractLastWord_ValidString_ReturnsLastWord(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedWord string
	}{
		{
			name:         "single word",
			input:        "field",
			expectedWord: "field",
		},
		{
			name:         "multiple words",
			input:        "WHERE field",
			expectedWord: "field",
		},
		{
			name:         "with AND",
			input:        "WHERE field1 = ? AND field2",
			expectedWord: "field2",
		},
		{
			name:         "with spaces",
			input:        "   field   ",
			expectedWord: "field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			word := extractLastWord(tt.input)
			assert.Equal(t, tt.expectedWord, word)
		})
	}
}

func TestExtractLastWord_EmptyString_ReturnsEmpty(t *testing.T) {
	word := extractLastWord("")
	assert.Empty(t, word)
}
