/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

// Package main provides the MCP server for EVA Team API.
//
// This CLI application implements the Model Context Protocol (MCP)
// for interacting with EVA Team project management system.
//
// Configuration can be provided via command-line flags or environment variables.
// Flags take precedence over environment variables.
//
// Flags:
//
//	--api-url, -u         EVA Team API URL (env: EVA_API_URL) [required]
//	--token, -t           API authentication token (env: EVA_API_TOKEN) [required]
//	--debug, -d           Enable debug logging for API requests (env: EVA_DEBUG)
//	--timeout             API request timeout (env: EVA_TIMEOUT) [default: 30s]
//	--transport           Transport: stdio or http (env: MCP_TRANSPORT) [default: stdio]
//	--http-addr           HTTP listen address (env: MCP_HTTP_ADDR) [default: 127.0.0.1:8080]
//	--http-path           HTTP base path for MCP endpoint (env: MCP_HTTP_PATH) [default: /mcp]
//	--http-stateless      Run HTTP transport in stateless mode (env: MCP_HTTP_STATELESS)
//	--http-json-response  Return application/json instead of SSE (env: MCP_HTTP_JSON_RESPONSE)
//
// Usage (stdio, for Claude Desktop / Claude Code CLI):
//
//	evateamclient-mcp --api-url="https://eva.example.com" --token="your-api-token"
//
// Usage (HTTP, for Claude.ai custom connector via tunnel):
//
//	evateamclient-mcp --transport=http --http-addr=127.0.0.1:8080 \
//	    --api-url="https://eva.example.com" --token="your-api-token"
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/raoptimus/evateamclient.go"
	"github.com/raoptimus/evateamclient.go/pkg/evateamclient-mcp/tools"
	"github.com/raoptimus/evateamclient.go/slogadapter"
	"github.com/urfave/cli/v3"
)

const (
	serverName            = "evateamclient-mcp"
	serverVersion         = "1.0.0"
	defaultRequestTimeout = 30 * time.Second
	defaultHTTPAddr       = "127.0.0.1:8080"
	defaultHTTPPath       = "/mcp"
	transportStdio        = "stdio"
	transportHTTP         = "http"
)

type transportConfig struct {
	Transport        string
	HTTPAddr         string
	HTTPPath         string
	HTTPStateless    bool
	HTTPJSONResponse bool
}

func main() {
	cfg := &evateamclient.Config{}
	tcfg := &transportConfig{}

	cmd := &cli.Command{
		Name:    serverName,
		Usage:   "MCP server for EVA Team API",
		Version: serverVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "api-url",
				Aliases:     []string{"u"},
				Usage:       "EVA Team API URL",
				Sources:     cli.EnvVars("EVA_API_URL"),
				Required:    true,
				Destination: &cfg.BaseURL,
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t"},
				Usage:       "API authentication token",
				Sources:     cli.EnvVars("EVA_API_TOKEN"),
				Required:    true,
				Destination: &cfg.APIToken,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "Enable debug logging for API requests and MCP server",
				Sources:     cli.EnvVars("EVA_DEBUG"),
				Destination: &cfg.Debug,
			},
			&cli.DurationFlag{
				Name:        "timeout",
				Usage:       "API request timeout",
				Sources:     cli.EnvVars("EVA_TIMEOUT"),
				Value:       defaultRequestTimeout,
				Destination: &cfg.Timeout,
			},
			&cli.StringFlag{
				Name:        "transport",
				Usage:       "Transport: stdio or http",
				Sources:     cli.EnvVars("MCP_TRANSPORT"),
				Value:       transportStdio,
				Destination: &tcfg.Transport,
			},
			&cli.StringFlag{
				Name:        "http-addr",
				Usage:       "HTTP listen address (used with --transport=http)",
				Sources:     cli.EnvVars("MCP_HTTP_ADDR"),
				Value:       defaultHTTPAddr,
				Destination: &tcfg.HTTPAddr,
			},
			&cli.StringFlag{
				Name:        "http-path",
				Usage:       "HTTP base path for MCP endpoint (used with --transport=http)",
				Sources:     cli.EnvVars("MCP_HTTP_PATH"),
				Value:       defaultHTTPPath,
				Destination: &tcfg.HTTPPath,
			},
			&cli.BoolFlag{
				Name:        "http-stateless",
				Usage:       "Run HTTP transport in stateless mode (used with --transport=http)",
				Sources:     cli.EnvVars("MCP_HTTP_STATELESS"),
				Destination: &tcfg.HTTPStateless,
			},
			&cli.BoolFlag{
				Name:        "http-json-response",
				Usage:       "Return application/json instead of SSE (used with --transport=http)",
				Sources:     cli.EnvVars("MCP_HTTP_JSON_RESPONSE"),
				Destination: &tcfg.HTTPJSONResponse,
			},
		},
		Writer:    os.Stderr,
		ErrWriter: os.Stderr,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(ctx, cfg, tcfg)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func runServer(ctx context.Context, cfg *evateamclient.Config, tcfg *transportConfig) error {
	// Create logger
	loggerLevel := slog.LevelWarn
	if cfg.Debug {
		loggerLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: loggerLevel,
	}))

	// Create EVA Team client
	evaClient, err := evateamclient.NewClient(cfg, evateamclient.WithLogger(slogadapter.New(logger)))
	if err != nil {
		return fmt.Errorf("failed to create EVA client: %w", err)
	}
	defer evaClient.Close()

	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    serverName,
			Version: serverVersion,
		},
		&mcp.ServerOptions{
			Logger: logger,
			Capabilities: &mcp.ServerCapabilities{
				Tools: &mcp.ToolCapabilities{ListChanged: true},
			},
		},
	)

	// Register tools before starting the server
	registry := tools.NewRegistry(evaClient)
	registry.RegisterAll(server)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	switch tcfg.Transport {
	case transportStdio, "":
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	case transportHTTP:
		return runHTTPServer(ctx, logger, server, httpOpts{
			Addr:         tcfg.HTTPAddr,
			Path:         tcfg.HTTPPath,
			Stateless:    tcfg.HTTPStateless,
			JSONResponse: tcfg.HTTPJSONResponse,
		})
	default:
		return fmt.Errorf("unknown transport %q (use %q or %q)", tcfg.Transport, transportStdio, transportHTTP)
	}
}
