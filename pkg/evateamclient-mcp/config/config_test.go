package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/raoptimus/evateamclient/pkg/evateamclient-mcp/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv_Success(t *testing.T) {
	// Setup
	os.Setenv("EVA_API_URL", "https://eva.example.com")
	os.Setenv("EVA_API_TOKEN", "test-token")
	os.Setenv("EVA_DEBUG", "true")
	os.Setenv("MCP_DEBUG", "true")
	os.Setenv("EVA_TIMEOUT", "60")
	defer func() {
		os.Unsetenv("EVA_API_URL")
		os.Unsetenv("EVA_API_TOKEN")
		os.Unsetenv("EVA_DEBUG")
		os.Unsetenv("MCP_DEBUG")
		os.Unsetenv("EVA_TIMEOUT")
	}()

	// Execute
	cfg, err := config.LoadFromEnv()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, "https://eva.example.com", cfg.APIURL)
	assert.Equal(t, "test-token", cfg.APIToken)
	assert.True(t, cfg.Debug)
	assert.True(t, cfg.MCPDebug)
	assert.Equal(t, 60*time.Second, cfg.Timeout)
}

func TestLoadFromEnv_MissingAPIURL_ReturnsError(t *testing.T) {
	// Setup
	os.Unsetenv("EVA_API_URL")
	os.Setenv("EVA_API_TOKEN", "test-token")
	defer os.Unsetenv("EVA_API_TOKEN")

	// Execute
	cfg, err := config.LoadFromEnv()

	// Verify
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "EVA_API_URL")
}

func TestLoadFromEnv_MissingAPIToken_ReturnsError(t *testing.T) {
	// Setup
	os.Setenv("EVA_API_URL", "https://eva.example.com")
	os.Unsetenv("EVA_API_TOKEN")
	defer os.Unsetenv("EVA_API_URL")

	// Execute
	cfg, err := config.LoadFromEnv()

	// Verify
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "EVA_API_TOKEN")
}

func TestLoadFromEnv_DefaultTimeout(t *testing.T) {
	// Setup
	os.Setenv("EVA_API_URL", "https://eva.example.com")
	os.Setenv("EVA_API_TOKEN", "test-token")
	os.Unsetenv("EVA_TIMEOUT")
	defer func() {
		os.Unsetenv("EVA_API_URL")
		os.Unsetenv("EVA_API_TOKEN")
	}()

	// Execute
	cfg, err := config.LoadFromEnv()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

func TestLoadFromEnv_InvalidTimeout_UsesDefault(t *testing.T) {
	// Setup
	os.Setenv("EVA_API_URL", "https://eva.example.com")
	os.Setenv("EVA_API_TOKEN", "test-token")
	os.Setenv("EVA_TIMEOUT", "invalid")
	defer func() {
		os.Unsetenv("EVA_API_URL")
		os.Unsetenv("EVA_API_TOKEN")
		os.Unsetenv("EVA_TIMEOUT")
	}()

	// Execute
	cfg, err := config.LoadFromEnv()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}
