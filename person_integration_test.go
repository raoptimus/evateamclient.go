package evateamclient

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_PersonsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityPerson).
		Limit(5)

	persons, meta, err := c.PersonsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, persons, "should have persons")

	for _, p := range persons {
		assert.True(t, strings.HasPrefix(p.ID, "CmfPerson:"), "ID should start with CmfPerson:, got %s", p.ID)
		assert.NotEmpty(t, p.Name)
	}
}

func TestIntegration_PersonsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultPersonListFields...).
		From(EntityPerson).
		Limit(5)

	persons, meta, err := c.PersonsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, persons)

	for _, p := range persons {
		assert.True(t, strings.HasPrefix(p.ID, "CmfPerson:"))
		assert.NotEmpty(t, p.Name)
		assert.NotEmpty(t, p.Code)
		assert.NotEmpty(t, p.Login)
	}
}

func TestIntegration_Person_ByID(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	// Get a person ID from the list first
	qb := NewQueryBuilder().
		Select(PersonFieldID).
		From(EntityPerson).
		Limit(1)

	persons, _, err := c.PersonsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, persons, "need at least one person")

	personID := persons[0].ID
	require.NotEmpty(t, personID)
	require.True(t, strings.HasPrefix(personID, "CmfPerson:"))

	// Fetch single person by ID with default fields
	person, meta, err := c.Person(ctx, personID, DefaultPersonFields)
	require.NoError(t, err)
	require.NotNil(t, person)
	require.NotNil(t, meta)

	assert.Equal(t, personID, person.ID)
	assert.NotEmpty(t, person.Name)
	assert.NotEmpty(t, person.Login)
	require.NotNil(t, person.CreatedAt, "CreatedAt should not be nil")
	assert.False(t, person.CreatedAt.IsZero(), "CreatedAt should not be zero")
}

func TestIntegration_Person_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	// Get a person ID from the list
	qb := NewQueryBuilder().
		Select(PersonFieldID).
		From(EntityPerson).
		Limit(1)

	persons, _, err := c.PersonsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, persons)

	// Fetch with ** (all basic + relation fields)
	person, meta, err := c.Person(ctx, persons[0].ID, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, person)
	require.NotNil(t, meta)

	assert.NotEmpty(t, person.ID)
	assert.NotEmpty(t, person.Name)
}

func TestIntegration_PersonQuery_WithBuilder(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(
			PersonFieldID,
			PersonFieldName,
			PersonFieldCode,
			PersonFieldLogin,
			PersonFieldEmail,
			PersonFieldCreatedAt,
		).
		From(EntityPerson).
		Limit(1)

	person, meta, err := c.PersonQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, person)
	require.NotNil(t, meta)

	assert.True(t, strings.HasPrefix(person.ID, "CmfPerson:"))
	assert.NotEmpty(t, person.Name)
	assert.NotEmpty(t, person.Code)
	assert.NotEmpty(t, person.Login)
	require.NotNil(t, person.CreatedAt)
	assert.False(t, person.CreatedAt.IsZero())
}

func TestIntegration_PersonCount(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(PersonFieldID).
		From(EntityPerson)

	count, err := c.PersonCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "should have persons")
}

func TestIntegration_Persons_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	ctx := context.Background()

	kwargs := map[string]any{
		"slice": []int{0, 5},
	}

	persons, meta, err := c.Persons(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, persons)

	for _, p := range persons {
		assert.True(t, strings.HasPrefix(p.ID, "CmfPerson:"))
		assert.NotEmpty(t, p.Name)
	}
}
