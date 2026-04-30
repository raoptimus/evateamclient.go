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

func TestClient_LogicTypeList_Success_ReturnsItems(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfLogicType:task.epic", "class_name": "CmfLogicType", "code": "task.epic", "name": "Epic", "cmf_model_name": "CmfTask"},
			{"id": "CmfLogicType:task.story", "class_name": "CmfLogicType", "code": "task.story", "name": "Story", "cmf_model_name": "CmfTask"}
		],
		"meta": {"total": 2}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfLogicType.list")
	}

	qb := NewQueryBuilder().From(EntityLogicType)
	items, meta, err := client.LogicTypeList(testCtx, qb)

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "task.epic", items[0].Code)
	assert.Equal(t, "Epic", items[0].Name)
	assert.Equal(t, "task.story", items[1].Code)
	assert.NotNil(t, meta)
}

func TestClient_LogicTypeByCode_Found_ReturnsLogicType(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [
			{"id": "CmfLogicType:task.epic", "class_name": "CmfLogicType", "code": "task.epic", "name": "Epic", "cmf_model_name": "CmfTask"}
		],
		"meta": {"total": 1}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)
	mockHTTP.urlCheck = func(url string) bool {
		return assert.Contains(t, url, "m=CmfLogicType.list")
	}

	lt, err := client.LogicTypeByCode(testCtx, LogicTypeEpic)

	require.NoError(t, err)
	require.NotNil(t, lt)
	assert.Equal(t, "CmfLogicType:task.epic", lt.ID)
	assert.Equal(t, LogicTypeEpic, lt.Code)
}

func TestClient_LogicTypeByCode_NotFound_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"result": [],
		"meta": {"total": 0}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	lt, err := client.LogicTypeByCode(testCtx, "task.unknown")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `logic type with code "task.unknown" not found`)
	assert.Nil(t, lt)
}

func TestClient_LogicTypeList_RPCError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	respBody := `{
		"jsonrpc": "2.2",
		"error": {"code": -32000, "message": "internal error"}
	}`

	mockHTTP.response = mockResponse(http.StatusOK, respBody)

	qb := NewQueryBuilder().From(EntityLogicType)
	items, meta, err := client.LogicTypeList(testCtx, qb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
	assert.Nil(t, items)
	assert.Nil(t, meta)
}

func TestClient_LogicTypeList_HTTPError_ReturnsError(t *testing.T) {
	client, mockHTTP := newTestClient(t)

	mockHTTP.err = errors.New("connection refused")

	qb := NewQueryBuilder().From(EntityLogicType)
	items, meta, err := client.LogicTypeList(testCtx, qb)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
	assert.Nil(t, items)
	assert.Nil(t, meta)
}
