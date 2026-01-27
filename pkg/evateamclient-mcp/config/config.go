// Package config provides configuration loading from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the MCP server.
type Config struct {
	// EVA API configuration
	APIURL   string
	APIToken string
	Debug    bool
	Timeout  time.Duration

	// MCP configuration
	MCPDebug bool
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Timeout: 30 * time.Second,
	}

	// Required: EVA_API_URL
	cfg.APIURL = os.Getenv("EVA_API_URL")
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("EVA_API_URL environment variable is required")
	}

	// Required: EVA_API_TOKEN
	cfg.APIToken = os.Getenv("EVA_API_TOKEN")
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("EVA_API_TOKEN environment variable is required")
	}

	// Optional: EVA_DEBUG
	if debugStr := os.Getenv("EVA_DEBUG"); debugStr != "" {
		cfg.Debug, _ = strconv.ParseBool(debugStr)
	}

	// Optional: MCP_DEBUG
	if debugStr := os.Getenv("MCP_DEBUG"); debugStr != "" {
		cfg.MCPDebug, _ = strconv.ParseBool(debugStr)
	}

	// Optional: EVA_TIMEOUT (in seconds)
	if timeoutStr := os.Getenv("EVA_TIMEOUT"); timeoutStr != "" {
		if seconds, err := strconv.Atoi(timeoutStr); err == nil && seconds > 0 {
			cfg.Timeout = time.Duration(seconds) * time.Second
		}
	}

	return cfg, nil
}
