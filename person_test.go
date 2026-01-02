package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Person_Success_ReturnsPerson(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": {
			"id": "CmfPerson:123",
			"name": "John Doe",
			"login": "john.doe",
			"email": "john@example.com"
		},
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfPerson.get")
	}

	person, meta, err := client.Person(testCtx, "CmfPerson:123", nil)

	require.NoError(t, err)
	require.NotNil(t, person)
	assert.Equal(t, "CmfPerson:123", person.ID)
	assert.Equal(t, "John Doe", person.Name)
	assert.Equal(t, "john.doe", person.Login)
	assert.NotNil(t, meta)
}

func TestClient_Person_NotFound_ReturnsRPCError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {
			"code": -32000,
			"message": "Person not found"
		}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	person, meta, err := client.Person(testCtx, "NONEXISTENT", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Person not found")
	assert.Nil(t, person)
	assert.Nil(t, meta)
}

func TestClient_Person_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	person, meta, err := client.Person(testCtx, "CmfPerson:123", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, person)
	assert.Nil(t, meta)
}

func TestClient_PersonsList_Success_ReturnsPersons(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfPerson:1", "name": "John Doe", "login": "john.doe"},
			{"id": "CmfPerson:2", "name": "Jane Smith", "login": "jane.smith"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfPerson.list")
	}

	qb := NewQueryBuilder().From(EntityPerson)
	persons, meta, err := client.PersonsList(testCtx, qb)

	require.NoError(t, err)
	assert.Len(t, persons, 2)
	assert.Equal(t, "john.doe", persons[0].Login)
	assert.Equal(t, "jane.smith", persons[1].Login)
	assert.NotNil(t, meta)
}

func TestClient_PersonCount_Success_ReturnsCount(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": 100
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfPerson.count")
	}

	qb := NewQueryBuilder().From(EntityPerson)
	count, err := client.PersonCount(testCtx, qb)

	require.NoError(t, err)
	assert.Equal(t, 100, count)
}

func TestClient_Persons_Success_ReturnsPersons(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfPerson:1", "name": "John Doe", "login": "john.doe"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfPerson.list")
	}

	persons, meta, err := client.Persons(testCtx, nil)

	require.NoError(t, err)
	assert.Len(t, persons, 1)
	assert.NotNil(t, meta)
}
