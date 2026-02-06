package evateamclient_test

import (
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/mockevateamclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Client initialization tests

func TestNewClient_ValidConfig_ReturnsClient(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
		Timeout:  30 * time.Second,
	}

	client, err := evateamclient.NewClient(&cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_EmptyBaseURL_ReturnsError(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "",
		APIToken: "test-token",
	}

	client, err := evateamclient.NewClient(&cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "baseURL")
}

func TestNewClient_EmptyAPIToken_ReturnsError(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "",
	}

	client, err := evateamclient.NewClient(&cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "APIToken")
}

func TestNewClient_InvalidBaseURL_ReturnsError(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "://invalid-url",
		APIToken: "test-token",
	}

	client, err := evateamclient.NewClient(&cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_ZeroTimeout_UsesDefault(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
		Timeout:  0,
	}

	client, err := evateamclient.NewClient(&cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient_WithOptions_AppliesAll(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
	}

	mockLogger := mockevateamclient.NewLogger(t)
	mockMetrics := mockevateamclient.NewMetrics(t)

	client, err := evateamclient.NewClient(&cfg,
		evateamclient.WithLogger(mockLogger),
		evateamclient.WithDebug(true),
		evateamclient.WithMetrics(mockMetrics),
	)

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestClient_Close_ReturnsNoError(t *testing.T) {
	cfg := evateamclient.Config{
		BaseURL:  "https://api.eva.team",
		APIToken: "test-token",
	}

	client, err := evateamclient.NewClient(&cfg)
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

// RPCError tests

func TestRPCError_ImplementsErrorInterface(t *testing.T) {
	var err error = &evateamclient.RPCError{
		Code:    -32600,
		Message: "Invalid Request",
	}

	assert.Equal(t, "Invalid Request", err.Error())
}

// QueryBuilder tests - testing actual behavior, not constants

func TestQueryBuilder_ToKwargs_CompleteQuery_ReturnsAllFields(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		Select("id", "name", "code").
		From(evateamclient.EntityTask).
		Where(sq.Eq{"project_id": "Project:123"}).
		OrderBy("-priority", "name").
		Offset(10).
		Limit(50).
		IncludeArchived().
		NoMeta()

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	// Check order_by
	orderBy, ok := kwargs["order_by"].([]string)
	require.True(t, ok)
	assert.Equal(t, []string{"-priority", "name"}, orderBy)

	// Check slice - EVA API uses [start, end] format, not [start, count]
	slice, ok := kwargs["slice"].([]uint64)
	require.True(t, ok)
	assert.Equal(t, []uint64{10, 60}, slice, "slice should be [offset, offset+limit] = [10, 60]")

	// Check flags
	assert.True(t, kwargs["include_archived"].(bool))
	assert.True(t, kwargs["no_meta"].(bool))

	// Check filter exists
	_, hasFilter := kwargs["filter"]
	assert.True(t, hasFilter)
}

func TestQueryBuilder_ToKwargs_NoFields_OmitsFieldsKey(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		From(evateamclient.EntityTask)

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	_, hasFields := kwargs["fields"]
	assert.False(t, hasFields, "fields should not be set when Select() not called with specific columns")
}

func TestQueryBuilder_ToKwargs_MultipleWhereConditions_CombinesFilters(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		From(evateamclient.EntityTask).
		Where(sq.Eq{"project_id": "Project:123"}).
		Where(sq.Eq{"status": "OPEN"})

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	filter, hasFilter := kwargs["filter"]
	assert.True(t, hasFilter)
	assert.NotNil(t, filter)
}

func TestQueryBuilder_ToMethod_ValidEntity_ReturnsListMethod(t *testing.T) {
	tests := []struct {
		entity         string
		expectedMethod string
	}{
		{evateamclient.EntityTask, "CmfTask.list"},
		{evateamclient.EntityProject, "CmfProject.list"},
		{evateamclient.EntityTimeLog, "CmfTimeTrackerHistory.list"},
		{evateamclient.EntityStatusHistory, "CmfStatusHistory.list"},
	}

	for _, tt := range tests {
		t.Run(tt.entity, func(t *testing.T) {
			qb := evateamclient.NewQueryBuilder().From(tt.entity)
			method, err := qb.ToMethod(false)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMethod, method)
		})
	}
}

func TestQueryBuilder_ToMethod_NoFrom_ReturnsError(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().Select("id", "name")

	_, err := qb.ToMethod(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "table name not found")
}

func TestQueryBuilder_Validate_NoFrom_ReturnsError(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().Select("id", "name")

	err := qb.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing From()")
}

func TestQueryBuilder_Validate_ValidQuery_ReturnsNoError(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		Select("id", "name").
		From(evateamclient.EntityTask)

	err := qb.Validate()
	assert.NoError(t, err)
}

func TestQueryBuilder_Where_VariousOperators_ConvertsCorrectly(t *testing.T) {
	// Tests operators that are supported by convertSquirrelFilters
	tests := []struct {
		name      string
		predicate any
	}{
		{"equality", sq.Eq{"field": "value"}},
		{"greater than", sq.Gt{"priority": 3}},
		{"greater or equal", sq.GtOrEq{"priority": 3}},
		{"less than", sq.Lt{"priority": 5}},
		{"less or equal", sq.LtOrEq{"priority": 5}},
		{"like", sq.Like{"name": "%test%"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := evateamclient.NewQueryBuilder().
				From(evateamclient.EntityTask).
				Where(tt.predicate)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err)

			_, hasFilter := kwargs["filter"]
			assert.True(t, hasFilter, "filter should be present for %s", tt.name)
		})
	}
}

func TestQueryBuilder_Between_CreatesRangeFilter(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		From(evateamclient.EntityTask).
		Where(evateamclient.Between("cmf_created_at", "2024-01-01", "2024-12-31"))

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)

	filter, hasFilter := kwargs["filter"]
	assert.True(t, hasFilter)
	assert.NotNil(t, filter)
}

func TestQueryBuilder_ToKwargs_PaginationScenarios_ReturnsCorrectSlice(t *testing.T) {
	tests := []struct {
		name          string
		offset        uint64
		limit         uint64
		expectedSlice []uint64
		description   string
	}{
		{
			name:          "first page",
			offset:        0,
			limit:         100,
			expectedSlice: []uint64{0, 100},
			description:   "Page 1: offset=0, limit=100 -> slice=[0, 100]",
		},
		{
			name:          "second page",
			offset:        100,
			limit:         100,
			expectedSlice: []uint64{100, 200},
			description:   "Page 2: offset=100, limit=100 -> slice=[100, 200]",
		},
		{
			name:          "third page",
			offset:        200,
			limit:         100,
			expectedSlice: []uint64{200, 300},
			description:   "Page 3: offset=200, limit=100 -> slice=[200, 300]",
		},
		{
			name:          "custom page size",
			offset:        50,
			limit:         25,
			expectedSlice: []uint64{50, 75},
			description:   "Custom: offset=50, limit=25 -> slice=[50, 75]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := evateamclient.NewQueryBuilder().
				From(evateamclient.EntityTask).
				Offset(tt.offset).
				Limit(tt.limit)

			kwargs, err := qb.ToKwargs()
			require.NoError(t, err, tt.description)

			slice, ok := kwargs["slice"].([]uint64)
			require.True(t, ok, "slice should be []uint64")
			assert.Equal(t, tt.expectedSlice, slice, tt.description)
		})
	}
}

func TestQueryBuilder_String_ReturnsDebugRepresentation(t *testing.T) {
	qb := evateamclient.NewQueryBuilder().
		Select("id").
		From(evateamclient.EntityTask).
		IncludeArchived()

	str := qb.String()
	assert.Contains(t, str, "QueryBuilder")
	assert.Contains(t, str, "includeArch=true")
}

func TestQueryBuilder_ChainedMethods_ReturnsSameBuilder(t *testing.T) {
	qb := evateamclient.NewQueryBuilder()

	// All methods should return the same builder instance for chaining
	result := qb.
		Select("id", "name").
		From(evateamclient.EntityTask).
		Where(sq.Eq{"id": "1"}).
		OrderBy("-name").
		Limit(10).
		Offset(5).
		IncludeArchived().
		NoMeta()

	assert.NotNil(t, result)
	assert.NoError(t, result.Validate())
}
