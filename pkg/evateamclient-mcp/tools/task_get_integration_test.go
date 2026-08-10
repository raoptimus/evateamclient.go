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
	"errors"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TaskGet_MissingCodeReturnsError verifies the E1 fix: the MCP
// eva_task_get handler surfaces a not-found error for a non-existent code,
// instead of returning an empty task object the way the raw EVA API does.
func TestIntegration_TaskGet_MissingCodeReturnsError(t *testing.T) {
	ctx := context.Background()

	client := newIntegrationClient(t)

	tt := tools.NewTaskTools(client)

	_, err := tt.TaskGet(ctx, &tools.TaskGetInput{
		Code:   "NOPE-00000000",
		Fields: tools.StringList{evateamclient.TaskFieldCode, evateamclient.TaskFieldName},
	})
	require.Error(t, err, "task_get for a non-existent code must return an error")
	assert.True(t, errors.Is(err, tools.ErrNotFound), "error must wrap ErrNotFound, got: %v", err)
}
