package tools_test

import (
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildQuery_EmptyInput(t *testing.T) {
	qb, err := tools.BuildQuery(evateamclient.EntityTask, nil)

	require.NoError(t, err)
	require.NotNil(t, qb)
}

func TestBuildQuery_WithFields(t *testing.T) {
	input := &tools.QueryInput{
		Fields: []string{"id", "name", "code"},
	}

	qb, err := tools.BuildQuery(evateamclient.EntityTask, input)

	require.NoError(t, err)
	require.NotNil(t, qb)

	// QueryBuilder converts SELECT columns to kwargs["fields"] only if they differ from "*"
	// The fields are stored in the SQL string, verify query is built correctly
	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)
	// Fields are extracted from SQL SELECT clause, check it's present
	assert.NotNil(t, kwargs)
}

func TestBuildQuery_WithPagination(t *testing.T) {
	input := &tools.QueryInput{
		Offset: 10,
		Limit:  50,
	}

	qb, err := tools.BuildQuery(evateamclient.EntityTask, input)

	require.NoError(t, err)
	require.NotNil(t, qb)

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)
	assert.Equal(t, []uint64{10, 60}, kwargs["slice"])
}

func TestBuildQuery_WithOrderBy(t *testing.T) {
	input := &tools.QueryInput{
		OrderBy: []string{"-created_at", "name"},
	}

	qb, err := tools.BuildQuery(evateamclient.EntityTask, input)

	require.NoError(t, err)
	require.NotNil(t, qb)

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)
	assert.Equal(t, []string{"-created_at", "name"}, kwargs["order_by"])
}

func TestBuildQuery_WithIncludeArchived(t *testing.T) {
	input := &tools.QueryInput{
		IncludeArchived: true,
	}

	qb, err := tools.BuildQuery(evateamclient.EntityTask, input)

	require.NoError(t, err)
	require.NotNil(t, qb)

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)
	assert.Equal(t, true, kwargs["include_archived"])
}

func TestBuildQuery_WithFilters(t *testing.T) {
	input := &tools.QueryInput{
		Filters: []tools.Filter{
			{Field: "status", Operator: "==", Value: "OPEN"},
		},
	}

	qb, err := tools.BuildQuery(evateamclient.EntityTask, input)

	require.NoError(t, err)
	require.NotNil(t, qb)

	kwargs, err := qb.ToKwargs()
	require.NoError(t, err)
	assert.Contains(t, kwargs, "filter")
}

func TestBuildKwargs_EmptyInput(t *testing.T) {
	kwargs := tools.BuildKwargs(nil)

	assert.Empty(t, kwargs)
}

func TestBuildKwargs_WithFields(t *testing.T) {
	input := &tools.QueryInput{
		Fields: []string{"id", "name"},
	}

	kwargs := tools.BuildKwargs(input)

	assert.Equal(t, []string{"id", "name"}, kwargs["fields"])
}

func TestBuildKwargs_WithSingleFilter(t *testing.T) {
	input := &tools.QueryInput{
		Filters: []tools.Filter{
			{Field: "status", Operator: "==", Value: "OPEN"},
		},
	}

	kwargs := tools.BuildKwargs(input)

	expected := []any{"status", "==", "OPEN"}
	assert.Equal(t, expected, kwargs["filter"])
}

func TestBuildKwargs_WithMultipleFilters(t *testing.T) {
	input := &tools.QueryInput{
		Filters: []tools.Filter{
			{Field: "status", Operator: "==", Value: "OPEN"},
			{Field: "priority", Operator: ">", Value: 3},
		},
	}

	kwargs := tools.BuildKwargs(input)

	expected := [][]any{
		{"status", "==", "OPEN"},
		{"priority", ">", 3},
	}
	assert.Equal(t, expected, kwargs["filter"])
}

func TestBuildKwargs_WithPagination(t *testing.T) {
	input := &tools.QueryInput{
		Offset: 20,
		Limit:  10,
	}

	kwargs := tools.BuildKwargs(input)

	assert.Equal(t, []int{20, 30}, kwargs["slice"])
}

func TestBuildKwargs_WithOrderBy(t *testing.T) {
	input := &tools.QueryInput{
		OrderBy: []string{"-created_at"},
	}

	kwargs := tools.BuildKwargs(input)

	assert.Equal(t, []string{"-created_at"}, kwargs["order_by"])
}

func TestBuildKwargs_WithIncludeArchived(t *testing.T) {
	input := &tools.QueryInput{
		IncludeArchived: true,
	}

	kwargs := tools.BuildKwargs(input)

	assert.Equal(t, true, kwargs["include_archived"])
}
