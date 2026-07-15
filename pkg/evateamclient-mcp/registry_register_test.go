/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package main

import (
	"testing"

	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/stretchr/testify/require"
)

// TestRegisterAll_BuildsRelaxedSchemasForEveryTool guards that registering every
// tool — which builds a relaxed input schema per tool via reflection — does not
// panic. Registration never calls the handlers, so a nil client is fine.
func TestRegisterAll_BuildsRelaxedSchemasForEveryTool(t *testing.T) {
	registry := tools.NewRegistry(nil)
	require.NotPanics(t, func() {
		registry.RegisterAll(newTestMCPServer())
	})
}
