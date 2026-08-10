/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTaskLinkServer creates an httptest.Server tailored for TaskLinkTools tests.
// The handler receives the parsed request kwargs and method name and returns the
// raw JSON to write as the response body. Returning "" causes a 400 response so
// the caller can detect unexpected calls. The returned *int counts every HTTP
// request the server received, so tests can prove a validation failure never
// reached the network (asserting on the error alone doesn't: a rejected mock
// request also produces an error).
func newTaskLinkServer(
	t *testing.T,
	handler func(method string, kwargs map[string]any) (responseJSON string),
) (*tools.TaskLinkTools, *int) {
	t.Helper()

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		body, _ := io.ReadAll(r.Body)
		method := r.URL.Query().Get("m")

		var req struct {
			Kwargs map[string]any `json:"kwargs"`
		}
		_ = json.Unmarshal(body, &req)

		responseBody := handler(method, req.Kwargs)
		if responseBody == "" {
			http.Error(w, "unexpected call: "+method, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(srv.Close)

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  srv.URL,
		APIToken: "test-token",
	})
	require.NoError(t, err)
	return tools.NewTaskLinkTools(client), &calls
}

// TestTaskLinkCreate_DefaultsRelationType_WhenEmpty pins the MCP-level default:
// an empty relation_type falls back to evateamclient.RelationTypeLink rather
// than erroring, per SPEC item 4.
func TestTaskLinkCreate_DefaultsRelationType_WhenEmpty(t *testing.T) {
	var captured map[string]any
	tt, _ := newTaskLinkServer(t, func(method string, kwargs map[string]any) string {
		if method == "CmfRelationOption.create" {
			captured = kwargs
			return `{"jsonrpc":"2.2","result":{"id":"CmfRelationOption:1","code":"RLO-001"}}`
		}
		return ""
	})

	input := tools.TaskLinkCreateInput{
		SourceTaskID: "TSK-000001",
		TargetTaskID: "TSK-000002",
	}

	_, err := tt.TaskLinkCreate(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, evateamclient.RelationTypeLink, captured["relation_type"])
}

// TestTaskLinkCreate_RelationType_Passthrough verifies an explicit relation_type
// reaches CmfRelationOption.create unchanged.
func TestTaskLinkCreate_RelationType_Passthrough(t *testing.T) {
	var captured map[string]any
	tt, _ := newTaskLinkServer(t, func(method string, kwargs map[string]any) string {
		if method == "CmfRelationOption.create" {
			captured = kwargs
			return `{"jsonrpc":"2.2","result":{"id":"CmfRelationOption:1","code":"RLO-001"}}`
		}
		return ""
	})

	input := tools.TaskLinkCreateInput{
		SourceTaskID: "TSK-000001",
		TargetTaskID: "TSK-000002",
		RelationType: "custom.type",
	}

	_, err := tt.TaskLinkCreate(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "custom.type", captured["relation_type"])
	_, hasID := captured["id"]
	assert.False(t, hasID, "kwargs must not contain the old, wrong 'id' key")
}

// TestTaskLinkCreate_MissingSourceOrTarget_ReturnsErrorWithoutRequest ensures
// missing task ids fail fast, without a HTTP round-trip. Asserting on the
// error alone would not prove this: the handler below returns "" for any
// method, which also produces an error via the (would-be) HTTP call.
func TestTaskLinkCreate_MissingSourceOrTarget_ReturnsErrorWithoutRequest(t *testing.T) {
	tt, calls := newTaskLinkServer(t, func(string, map[string]any) string { return "" })

	_, err := tt.TaskLinkCreate(context.Background(), tools.TaskLinkCreateInput{TargetTaskID: "TSK-000002"})
	assert.Error(t, err)

	_, err = tt.TaskLinkCreate(context.Background(), tools.TaskLinkCreateInput{SourceTaskID: "TSK-000001"})
	assert.Error(t, err)

	assert.Equal(t, 0, *calls, "validation must fail before any HTTP request")
}
