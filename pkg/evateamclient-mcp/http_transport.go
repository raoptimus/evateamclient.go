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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	httpReadHeaderTimeout = 10 * time.Second
	httpShutdownTimeout   = 5 * time.Second
)

type httpOpts struct {
	Addr         string
	Path         string
	Stateless    bool
	JSONResponse bool
}

func runHTTPServer(ctx context.Context, logger *slog.Logger, server *mcp.Server, opts httpOpts) error {
	if opts.Addr == "" {
		return errors.New("http addr is required")
	}
	if opts.Path == "" {
		opts.Path = "/mcp"
	}

	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{
			Stateless:    opts.Stateless,
			JSONResponse: opts.JSONResponse,
			Logger:       logger,
		},
	)

	mux := http.NewServeMux()
	mux.Handle(opts.Path, mcpHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			logger.Debug("healthz write failed", "err", err)
		}
	})

	httpServer := &http.Server{
		Addr:              opts.Addr,
		Handler:           mux,
		ReadHeaderTimeout: httpReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("MCP HTTP server listening", "addr", opts.Addr, "path", opts.Path)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), httpShutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
		<-errCh
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	}
}
