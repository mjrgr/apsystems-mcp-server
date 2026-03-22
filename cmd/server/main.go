// APsystems MCP Server
//
// This binary starts an MCP server that exposes APsystems solar monitoring
// data as MCP tools. It communicates over stdio (JSON-RPC) by default, or
// over HTTP with Server-Sent Events (SSE) when configured. Optionally runs
// an HTTP dashboard for visual monitoring.
//
// Environment variables:
//
//	APS_APP_ID        – 32-character APsystems OpenAPI App ID (required)
//	APS_APP_SECRET    – 12-character APsystems OpenAPI App Secret (required)
//	APS_SYS_ID        – Default system ID (sid) used when not provided in tool arguments (optional)
//	APS_BASE_URL      – Override the API base URL (optional)
//	APS_MCP_TRANSPORT – MCP transport: "stdio" (default) or "sse"
//	APS_MCP_SSE_ADDR  – SSE server listen address, default ":8888" (only used when transport is "sse")
//	APS_DASHBOARD     – "true" to enable the web dashboard (optional)
//	APS_DASH_ADDR     – Dashboard listen address, default ":8080" (optional)
//	APS_LOG_LEVEL     – Log level: debug, info, warn, error (default: info)

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mjrgr/apsystems-mcp-server/internal/api"
	"github.com/mjrgr/apsystems-mcp-server/internal/config"
	"github.com/mjrgr/apsystems-mcp-server/internal/dashboard"
	"github.com/mjrgr/apsystems-mcp-server/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := newLogger(cfg.LogLevel)
	logger.Info("apsystems-mcp-server starting",
		"transport", cfg.Transport,
		"dashboard", cfg.Dashboard,
		"base_url", cfg.BaseURL,
	)

	opts := []api.Option{api.WithLogger(logger)}
	if cfg.BaseURL != "" {
		opts = append(opts, api.WithBaseURL(cfg.BaseURL))
	}
	client := api.NewClient(cfg.AppID, cfg.AppSecret, opts...)
	mcpServer := mcp.New(client, logger, cfg.SysID)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Dashboard {
		dash := dashboard.New(client, logger, cfg.SysID)
		go func() {
			if err := dash.Serve(ctx, cfg.DashAddr); err != nil {
				logger.Error("dashboard error", "err", err)
			}
		}()
		logger.Info("dashboard enabled", "addr", cfg.DashAddr)

		go func() {
			switch cfg.Transport {
			case "sse":
				if err := mcpServer.ServeSSE(ctx, cfg.SSEAddr); err != nil {
					logger.Error("mcp sse error", "err", err)
				}
			default:
				if err := mcpServer.ServeStdio(ctx); err != nil {
					logger.Error("mcp stdio error", "err", err)
				}
			}
		}()

		<-ctx.Done()
		return nil
	}

	switch cfg.Transport {
	case "sse":
		return mcpServer.ServeSSE(ctx, cfg.SSEAddr)
	default: // stdio
		return mcpServer.ServeStdio(ctx)
	}
}

// newLogger creates a structured logger at the requested level.
func newLogger(level slog.Level) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	// When using stdio transport the MCP protocol owns stdout, so log to stderr.
	return slog.New(slog.NewJSONHandler(os.Stderr, opts))
}
