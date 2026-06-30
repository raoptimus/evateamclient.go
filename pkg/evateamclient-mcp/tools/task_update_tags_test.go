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

func newTaskToolsWithServer(t *testing.T, capturedKwargs *map[string]any) *tools.TaskTools {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		m := r.URL.Query().Get("m")
		switch m {
		case "CmfTask.update":
			var req struct {
				Kwargs map[string]any `json:"kwargs"`
			}
			_ = json.Unmarshal(body, &req)
			if capturedKwargs != nil {
				*capturedKwargs = req.Kwargs
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.2","result":"CmfTask:test-id"}`))
		case "CmfTask.get":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.2","result":{"id":"CmfTask:test-id","class_name":"CmfTask","name":"Test"}}`))
		case "CmfTask.list":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.2","result":[{"id":"CmfTask:test-id","class_name":"CmfTask","name":"Test"}],"meta":{"total":1}}`))
		default:
			http.Error(w, "unknown method", http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  srv.URL,
		APIToken: "test",
	})
	require.NoError(t, err)
	return tools.NewTaskTools(client)
}

func TestTaskUpdate_Tags_ArrayPassthrough(t *testing.T) {
	var captured map[string]any
	tt := newTaskToolsWithServer(t, &captured)

	_, err := tt.TaskUpdate(context.Background(), tools.TaskUpdateInput{
		ID:      "CmfTask:test-id",
		Updates: map[string]any{"tags": []any{"TAG-001", "TAG-002"}},
	})

	require.NoError(t, err)
	assert.Equal(t, []any{"TAG-001", "TAG-002"}, captured["tags"])
}

func TestTaskUpdate_Tags_JSONEncodedString(t *testing.T) {
	var captured map[string]any
	tt := newTaskToolsWithServer(t, &captured)

	_, err := tt.TaskUpdate(context.Background(), tools.TaskUpdateInput{
		ID:      "CmfTask:test-id",
		Updates: map[string]any{"tags": `["TAG-004379","TAG-004400","TAG-004402"]`},
	})

	require.NoError(t, err)
	assert.Equal(t, []any{"TAG-004379", "TAG-004400", "TAG-004402"}, captured["tags"])
}

func TestTaskUpdate_Tags_Missing(t *testing.T) {
	var captured map[string]any
	tt := newTaskToolsWithServer(t, &captured)

	_, err := tt.TaskUpdate(context.Background(), tools.TaskUpdateInput{
		ID:      "CmfTask:test-id",
		Updates: map[string]any{"name": "New name"},
	})

	require.NoError(t, err)
	assert.Nil(t, captured["tags"])
	assert.Equal(t, "New name", captured["name"])
}

func TestTaskUpdate_Tags_InvalidStringRemoved(t *testing.T) {
	var captured map[string]any
	tt := newTaskToolsWithServer(t, &captured)

	_, err := tt.TaskUpdate(context.Background(), tools.TaskUpdateInput{
		ID:      "CmfTask:test-id",
		Updates: map[string]any{"tags": "not-valid-json"},
	})

	require.NoError(t, err)
	assert.Nil(t, captured["tags"])
}
