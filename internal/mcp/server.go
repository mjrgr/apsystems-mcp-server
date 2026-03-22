// Package mcp registers APsystems tools with an MCP server.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/mjrgr/apsystems-mcp-server/internal/api"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and the APsystems API client.
type Server struct {
	mcp    *server.MCPServer
	client *api.Client
	logger *slog.Logger
	sysID  string
}

// NewServer creates a new MCP server with all APsystems tools registered.
func NewServer(client *api.Client, logger *slog.Logger, sysID string) *Server {
	s := &Server{
		mcp: server.NewMCPServer(
			"APsystems Solar Monitor",
			"1.0.0",
			server.WithToolCapabilities(true),
		),
		client: client,
		logger: logger,
		sysID:  sysID,
	}
	s.registerTools()
	return s
}

// MCPServer returns the underlying MCP server for transport binding.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcp
}

// registerTools adds all APsystems tool definitions.
func (s *Server) registerTools() {
	// ── System Details ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_system_details",
			Description: "Get details for a particular APsystems solar system including capacity, timezone, ECU list, and status light.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier (e.g. AZ12649A3DFF)",
					},
				},
				Required: []string{"sid"},
			},
		},
		s.handleGetSystemDetails,
	)

	// ── System Inverters ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_inverters",
			Description: "List all ECUs and their connected micro-inverters for a system.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
				},
				Required: []string{"sid"},
			},
		},
		s.handleGetInverters,
	)

	// ── System Meters ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_meters",
			Description: "List all meter IDs for a system.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
				},
				Required: []string{"sid"},
			},
		},
		s.handleGetMeters,
	)

	// ── System Summary ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_system_summary",
			Description: "Get accumulative energy summary (today, month, year, lifetime) for a system in kWh.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
				},
				Required: []string{"sid"},
			},
		},
		s.handleGetSystemSummary,
	)

	// ── System Energy ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_system_energy",
			Description: "Get energy data for a system at different time granularities. Returns energy values in kWh.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "Time granularity: hourly, daily, monthly, or yearly",
						"enum":        []string{"hourly", "daily", "monthly", "yearly"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date filter. Format: yyyy-MM-dd (hourly), yyyy-MM (daily), yyyy (monthly), omit for yearly.",
					},
				},
				Required: []string{"sid", "energy_level"},
			},
		},
		s.handleGetSystemEnergy,
	)

	// ── ECU Summary ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_ecu_summary",
			Description: "Get accumulative energy summary for a particular ECU (Energy Communication Unit).",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The ECU identifier",
					},
				},
				Required: []string{"sid", "eid"},
			},
		},
		s.handleGetECUSummary,
	)

	// ── ECU Energy ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_ecu_energy",
			Description: "Get energy data for a particular ECU at different granularities. Supports minutely telemetry.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The ECU identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "Time granularity: minutely, hourly, daily, monthly, or yearly",
						"enum":        []string{"minutely", "hourly", "daily", "monthly", "yearly"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date filter. Format depends on energy_level.",
					},
				},
				Required: []string{"sid", "eid", "energy_level"},
			},
		},
		s.handleGetECUEnergy,
	)

	// ── Inverter Summary ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_inverter_summary",
			Description: "Get per-channel energy summary for a specific micro-inverter (today/month/year/lifetime per channel).",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"uid": map[string]interface{}{
						"type":        "string",
						"description": "The inverter identifier",
					},
				},
				Required: []string{"sid", "uid"},
			},
		},
		s.handleGetInverterSummary,
	)

	// ── Inverter Energy ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_inverter_energy",
			Description: "Get per-channel energy data for a specific micro-inverter at different granularities. Minutely level includes DC power/current/voltage and AC telemetry.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"uid": map[string]interface{}{
						"type":        "string",
						"description": "The inverter identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "Time granularity: minutely, hourly, daily, monthly, or yearly",
						"enum":        []string{"minutely", "hourly", "daily", "monthly", "yearly"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date filter. Format depends on energy_level.",
					},
				},
				Required: []string{"sid", "uid", "energy_level"},
			},
		},
		s.handleGetInverterEnergy,
	)

	// ── Inverter Batch Energy ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_inverter_batch_energy",
			Description: "Get energy or power telemetry for all inverters under an ECU in a single day.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The ECU identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "power (telemetry) or energy (daily totals)",
						"enum":        []string{"power", "energy"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date in yyyy-MM-dd format",
					},
				},
				Required: []string{"sid", "eid", "energy_level", "date_range"},
			},
		},
		s.handleGetInverterBatchEnergy,
	)

	// ── Meter Summary ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_meter_summary",
			Description: "Get energy summary for a meter (consumed, exported, imported, produced) by today/month/year/lifetime.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The meter ECU identifier",
					},
				},
				Required: []string{"sid", "eid"},
			},
		},
		s.handleGetMeterSummary,
	)

	// ── Meter Period ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_meter_period",
			Description: "Get period energy data for a meter at different granularities.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The meter ECU identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "Time granularity: minutely, hourly, daily, monthly, or yearly",
						"enum":        []string{"minutely", "hourly", "daily", "monthly", "yearly"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date filter. Format depends on energy_level.",
					},
				},
				Required: []string{"sid", "eid", "energy_level"},
			},
		},
		s.handleGetMeterPeriod,
	)

	// ── Storage Latest ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_storage_latest",
			Description: "Get the latest status and power readings for a storage ECU (battery SOC, charge/discharge power, etc.).",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The storage ECU identifier",
					},
				},
				Required: []string{"sid", "eid"},
			},
		},
		s.handleGetStorageLatest,
	)

	// ── Storage Summary ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_storage_summary",
			Description: "Get accumulative energy summary for a storage ECU (charge/discharge/produced/consumed/exported/imported).",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The storage ECU identifier",
					},
				},
				Required: []string{"sid", "eid"},
			},
		},
		s.handleGetStorageSummary,
	)

	// ── Storage Period ──
	s.mcp.AddTool(
		mcplib.Tool{
			Name:        "get_storage_period",
			Description: "Get period energy data for a storage ECU at different granularities.",
			InputSchema: mcplib.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"sid": map[string]interface{}{
						"type":        "string",
						"description": "The unique system identifier",
					},
					"eid": map[string]interface{}{
						"type":        "string",
						"description": "The storage ECU identifier",
					},
					"energy_level": map[string]interface{}{
						"type":        "string",
						"description": "Time granularity: minutely, hourly, daily, monthly, or yearly",
						"enum":        []string{"minutely", "hourly", "daily", "monthly", "yearly"},
					},
					"date_range": map[string]interface{}{
						"type":        "string",
						"description": "Date filter. Format depends on energy_level.",
					},
				},
				Required: []string{"sid", "eid", "energy_level"},
			},
		},
		s.handleGetStoragePeriod,
	)
}

// ─── Handler implementations ────────────────────────────────────────────────

func (s *Server) handleGetSystemDetails(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetSystemDetails(ctx, sid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetInverters(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetSystemInverters(ctx, sid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetMeters(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetSystemMeters(ctx, sid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetSystemSummary(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetSystemSummary(ctx, sid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetSystemEnergy(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange := optionalString(req, "date_range")

	data, err := s.client.GetSystemEnergy(ctx, sid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

func (s *Server) handleGetECUSummary(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetECUSummary(ctx, sid, eid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetECUEnergy(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange := optionalString(req, "date_range")

	data, err := s.client.GetECUEnergy(ctx, sid, eid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

func (s *Server) handleGetInverterSummary(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	uid, err := requireString(req, "uid", "")
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetInverterSummary(ctx, sid, uid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetInverterEnergy(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	uid, err := requireString(req, "uid", "")
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange := optionalString(req, "date_range")

	data, err := s.client.GetInverterEnergy(ctx, sid, uid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

func (s *Server) handleGetInverterBatchEnergy(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange, err := requireString(req, "date_range", "")
	if err != nil {
		return errResult(err), nil
	}

	data, err := s.client.GetInverterBatchEnergy(ctx, sid, eid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

func (s *Server) handleGetMeterSummary(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetMeterSummary(ctx, sid, eid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetMeterPeriod(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange := optionalString(req, "date_range")

	data, err := s.client.GetMeterPeriod(ctx, sid, eid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

func (s *Server) handleGetStorageLatest(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetStorageLatest(ctx, sid, eid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetStorageSummary(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	data, err := s.client.GetStorageSummary(ctx, sid, eid)
	if err != nil {
		return errResult(err), nil
	}
	return jsonResult(data)
}

func (s *Server) handleGetStoragePeriod(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sid, err := requireString(req, "sid", s.sysID)
	if err != nil {
		return errResult(err), nil
	}
	eid, err := requireString(req, "eid", "")
	if err != nil {
		return errResult(err), nil
	}
	level, err := requireString(req, "energy_level", "")
	if err != nil {
		return errResult(err), nil
	}
	dateRange := optionalString(req, "date_range")

	data, err := s.client.GetStoragePeriod(ctx, sid, eid, level, dateRange)
	if err != nil {
		return errResult(err), nil
	}
	return rawResult(data)
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func requireString(req mcplib.CallToolRequest, key, fallback string) (string, error) {
	v, ok := req.GetArguments()[key]
	if !ok {
		if fallback != "" {
			return fallback, nil
		}
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("parameter %s must be a non-empty string", key)
	}
	return s, nil
}

func optionalString(req mcplib.CallToolRequest, key string) string {
	v, ok := req.GetArguments()[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func jsonResult(v interface{}) (*mcplib.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errResult(err), nil
	}
	return mcplib.NewToolResultText(string(b)), nil
}

func rawResult(data json.RawMessage) (*mcplib.CallToolResult, error) {
	// Pretty-print the raw JSON.
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return mcplib.NewToolResultText(string(data)), nil
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return mcplib.NewToolResultText(string(b)), nil
}

func errResult(err error) *mcplib.CallToolResult {
	return mcplib.NewToolResultError(err.Error())
}
