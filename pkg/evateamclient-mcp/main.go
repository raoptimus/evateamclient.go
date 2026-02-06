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
//	--api-url, -u    EVA Team API URL (env: EVA_API_URL) [required]
//	--token, -t      API authentication token (env: EVA_API_TOKEN) [required]
//	--debug, -d      Enable debug logging for API requests (env: EVA_DEBUG)
//	--mcp-debug      Enable debug logging for MCP server (env: MCP_DEBUG)
//	--timeout        API request timeout (env: EVA_TIMEOUT) [default: 30s]
//
// Usage:
//
//	evateamclient-mcp --api-url="https://eva.example.com" --token="your-api-token"
//
// Or with environment variables:
//
//	export EVA_API_URL="https://eva.example.com"
//	export EVA_API_TOKEN="your-api-token"
//	evateamclient-mcp
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
)

func main() {
	cfg := &evateamclient.Config{}

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
		},
		Writer:    os.Stderr,
		ErrWriter: os.Stderr,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(ctx, cfg)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func runServer(ctx context.Context, cfg *evateamclient.Config) error {
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

	// Run server on stdio transport
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
