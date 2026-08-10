/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package tools_test

import (
	"os"
	"testing"

	evateamclient "github.com/raoptimus/evateamclient.go"
	"github.com/stretchr/testify/require"
)

// newIntegrationClient builds a client against the real EVA API, skipping the
// test when the credentials are absent — integration tests are gated by
// EVA_API_URL/EVA_API_TOKEN and must not fail a credential-less run.
func newIntegrationClient(t *testing.T) *evateamclient.Client {
	t.Helper()

	baseURL, token := os.Getenv("EVA_API_URL"), os.Getenv("EVA_API_TOKEN")
	if baseURL == "" || token == "" {
		t.Skip("EVA_API_URL and EVA_API_TOKEN are required for integration tests")
	}

	client, err := evateamclient.NewClient(&evateamclient.Config{
		BaseURL:  baseURL,
		APIToken: token,
		Debug:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	return client
}
