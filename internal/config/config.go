package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	// AppID is the 32-character APsystems OpenAPI App ID (required).
	AppID string
	// AppSecret is the 12-character APsystems OpenAPI App Secret (required).
	AppSecret string
	// SysID is the default system identifier used when not provided in tool arguments (optional).
	SysID string
	// BaseURL overrides the APsystems API base URL (optional).
	BaseURL string
	// Transport selects the MCP transport: "stdio" or "sse" (default: stdio).
	Transport string
	// SSEAddr is the listen address for the SSE HTTP server (default: :8888).
	SSEAddr string
	// Dashboard enables the web dashboard (default: false).
	Dashboard bool
	// DashAddr is the listen address for the web dashboard (default: :8080).
	DashAddr string
	// LogLevel controls log verbosity (default: info).
	LogLevel slog.Level
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	appID := os.Getenv("APS_APP_ID")
	if appID == "" {
		return nil, fmt.Errorf("APS_APP_ID environment variable is required")
	}
	appSecret := os.Getenv("APS_APP_SECRET")
	if appSecret == "" {
		return nil, fmt.Errorf("APS_APP_SECRET environment variable is required")
	}

	transport := "stdio"
	if t := strings.ToLower(os.Getenv("APS_MCP_TRANSPORT")); t != "" {
		switch t {
		case "stdio", "sse":
			transport = t
		default:
			return nil, fmt.Errorf("invalid APS_MCP_TRANSPORT %q: must be stdio or sse", t)
		}
	}

	sseAddr := ":8888"
	if a := os.Getenv("APS_MCP_SSE_ADDR"); a != "" {
		sseAddr = a
	} else if p := os.Getenv("APS_MCP_SSE_PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid APS_MCP_SSE_PORT %q: must be a number between 1 and 65535", p)
		}
		sseAddr = fmt.Sprintf(":%d", port)
	}

	dashAddr := ":8080"
	if a := os.Getenv("APS_DASH_ADDR"); a != "" {
		dashAddr = a
	}

	logLevel := slog.LevelInfo
	if ll := strings.ToLower(os.Getenv("APS_LOG_LEVEL")); ll != "" {
		switch ll {
		case "debug":
			logLevel = slog.LevelDebug
		case "info":
			logLevel = slog.LevelInfo
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			return nil, fmt.Errorf("invalid APS_LOG_LEVEL %q: must be debug, info, warn, or error", ll)
		}
	}

	return &Config{
		AppID:     appID,
		AppSecret: appSecret,
		SysID:     os.Getenv("APS_SYS_ID"),
		BaseURL:   os.Getenv("APS_BASE_URL"),
		Transport: transport,
		SSEAddr:   sseAddr,
		Dashboard: strings.EqualFold(os.Getenv("APS_DASHBOARD"), "true"),
		DashAddr:  dashAddr,
		LogLevel:  logLevel,
	}, nil
}
