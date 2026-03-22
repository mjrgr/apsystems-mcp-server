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
	"strings"
	"syscall"

	"github.com/mjrgr/apsystems-mcp-server/internal/api"
	"github.com/mjrgr/apsystems-mcp-server/internal/dashboard"
	mcpserver "github.com/mjrgr/apsystems-mcp-server/internal/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// ── Parse config from environment ──
	appID := requireEnv("APS_APP_ID")
	appSecret := requireEnv("APS_APP_SECRET")
	sysID := os.Getenv("APS_SYS_ID")

	logLevel := envOr("APS_LOG_LEVEL", "info")
	logger := newLogger(logLevel)

	// ── Build the API client ──
	opts := []api.Option{
		api.WithLogger(logger),
	}
	if base := os.Getenv("APS_BASE_URL"); base != "" {
		opts = append(opts, api.WithBaseURL(base))
	}
	client := api.NewClient(appID, appSecret, opts...)

	// ── Context with signal handling ──
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// ── Optionally start the web dashboard ──
	if strings.EqualFold(os.Getenv("APS_DASHBOARD"), "true") {
		addr := envOr("APS_DASH_ADDR", ":8080")
		dashSrv := dashboard.New(client, logger, sysID)
		go func() {
			if err := dashSrv.Serve(ctx, addr); err != nil {
				logger.Error("dashboard error", "err", err)
			}
		}()
	}

	// ── Start the MCP server ──
	mcpSrv := mcpserver.NewServer(client, logger, sysID)
	transport := strings.ToLower(envOr("APS_MCP_TRANSPORT", "stdio"))

	switch transport {
	case "sse":
		sseAddr := envOr("APS_MCP_SSE_ADDR", ":8888")
		sseServer := server.NewSSEServer(mcpSrv.MCPServer(),
			server.WithBaseURL("http://localhost"+sseAddr),
		)
		logger.Info("starting MCP server with SSE transport", "addr", sseAddr)

		go func() {
			<-ctx.Done()
			if err := sseServer.Shutdown(context.Background()); err != nil {
				logger.Error("sse server shutdown error", "err", err)
			}
		}()

		if err := sseServer.Start(sseAddr); err != nil {
			logger.Error("sse server error", "err", err)
			cancel()
			os.Exit(1)
		}
	default:
		logger.Info("starting MCP server on stdio")
		if err := server.ServeStdio(mcpSrv.MCPServer()); err != nil {
			logger.Error("mcp server error", "err", err)
			cancel()
			os.Exit(1)
		}
	}

	// Block until shutdown signal
	<-ctx.Done()
	cancel()
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "ERROR: required environment variable %s is not set\n", key)
		os.Exit(1)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newLogger(level string) *slog.Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
}
