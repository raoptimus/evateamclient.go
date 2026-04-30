/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMCPServer() *mcp.Server {
	return mcp.NewServer(
		&mcp.Implementation{Name: "test-mcp", Version: "0.0.1"},
		nil,
	)
}

func pickFreeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())
	return addr
}

func waitForServer(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server at %s did not start in time", url)
}

func TestRunHTTPServer_EmptyAddr_ReturnsError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := runHTTPServer(context.Background(), logger, newTestMCPServer(), httpOpts{Addr: ""})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "http addr is required")
}

func TestRunHTTPServer_StopsOnContextCancel_NoError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	addr := pickFreeAddr(t)

	ctx, cancel := context.WithCancel(t.Context())
	errCh := make(chan error, 1)
	go func() {
		errCh <- runHTTPServer(ctx, logger, newTestMCPServer(), httpOpts{
			Addr: addr, Path: "/mcp",
		})
	}()

	waitForServer(t, fmt.Sprintf("http://%s/healthz", addr))
	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(httpShutdownTimeout + time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestRunHTTPServer_HealthzEndpoint_ReturnsOK(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	addr := pickFreeAddr(t)

	go func() {
		_ = runHTTPServer(t.Context(), logger, newTestMCPServer(), httpOpts{
			Addr: addr, Path: "/mcp",
		})
	}()

	url := fmt.Sprintf("http://%s/healthz", addr)
	waitForServer(t, url)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(body))
}

func TestRunHTTPServer_DefaultPath_UsedWhenEmpty(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	addr := pickFreeAddr(t)

	go func() {
		_ = runHTTPServer(t.Context(), logger, newTestMCPServer(), httpOpts{
			Addr: addr, Path: "", // expect default /mcp
		})
	}()

	waitForServer(t, fmt.Sprintf("http://%s/healthz", addr))

	resp, err := http.Get(fmt.Sprintf("http://%s/mcp", addr))
	require.NoError(t, err)
	defer resp.Body.Close()

	// MCP handler must be registered at default path. GET without proper
	// headers is rejected by the handler, but it should not return 404.
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
}
