/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_TagList_Success_ReturnsItems(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTag:1", "class_name": "CmfTag", "name": "Backend", "code": "TAG-000001", "alias": ["Backend"]},
			{"id": "CmfTag:2", "class_name": "CmfTag", "name": "Frontend", "code": "TAG-000002", "alias": []}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfTag.list")
	}

	qb := NewQueryBuilder().From(EntityTag)
	items, meta, err := client.TagList(testCtx, qb)

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "CmfTag:1", items[0].ID)
	assert.Equal(t, "CmfTag", items[0].ClassName)
	assert.Equal(t, "Backend", items[0].Name)
	assert.Equal(t, "TAG-000001", items[0].Code)
	assert.Equal(t, []string{"Backend"}, items[0].Alias)
	assert.Equal(t, "TAG-000002", items[1].Code)
	assert.Equal(t, "Frontend", items[1].Name)
	assert.NotNil(t, meta)
}

func TestClient_TagList_DefaultFields_AppliedWhenNotSet(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfTag:3", "class_name": "CmfTag", "name": "DevOps", "code": "TAG-000003", "alias": []}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.bodyCheck = func(body []byte) bool {
		for _, field := range DefaultTagFields {
			if !assert.Contains(t, string(body), field) {
				return false
			}
		}
		return true
	}

	qb := NewQueryBuilder().From(EntityTag)
	items, meta, err := client.TagList(testCtx, qb)

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "TAG-000003", items[0].Code)
	assert.NotNil(t, meta)
}

func TestClient_TagList_RPCError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {"code": -32000, "message": "internal error"}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	qb := NewQueryBuilder().From(EntityTag)
	items, meta, err := client.TagList(testCtx, qb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
	assert.Nil(t, items)
	assert.Nil(t, meta)
}

func TestClient_TagList_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	qb := NewQueryBuilder().From(EntityTag)
	items, meta, err := client.TagList(testCtx, qb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, items)
	assert.Nil(t, meta)
}
