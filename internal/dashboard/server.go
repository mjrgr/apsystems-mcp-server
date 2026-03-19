// Package dashboard provides an optional HTTP dashboard for monitoring
// APsystems solar data visually. It serves a single-page app and proxies
// API requests to the APsystems backend.
package dashboard

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/apsystems/mcp-server/internal/api"
)

//go:embed dashboard.html
var dashboardHTML []byte

// Handler creates an http.Handler that serves the dashboard UI and
// proxied API endpoints.
func Handler(client *api.Client, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// ── API proxy routes ──
	mux.HandleFunc("GET /api/system/{sid}", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		data, err := client.GetSystemDetails(r.Context(), sid)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("GET /api/system/{sid}/summary", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		data, err := client.GetSystemSummary(r.Context(), sid)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("GET /api/system/{sid}/energy", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		level := r.URL.Query().Get("energy_level")
		dateRange := r.URL.Query().Get("date_range")
		if level == "" {
			level = "daily"
		}
		data, err := client.GetSystemEnergy(r.Context(), sid, level, dateRange)
		writeJSON(w, json.RawMessage(data), err)
	})

	mux.HandleFunc("GET /api/system/{sid}/inverters", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		data, err := client.GetSystemInverters(r.Context(), sid)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("GET /api/system/{sid}/ecu/{eid}/summary", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		data, err := client.GetECUSummary(r.Context(), sid, eid)
		writeJSON(w, data, err)
	})

	mux.HandleFunc("GET /api/system/{sid}/ecu/{eid}/energy", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		level := r.URL.Query().Get("energy_level")
		dateRange := r.URL.Query().Get("date_range")
		if level == "" {
			level = "daily"
		}
		data, err := client.GetECUEnergy(r.Context(), sid, eid, level, dateRange)
		writeJSON(w, json.RawMessage(data), err)
	})

	mux.HandleFunc("GET /api/system/{sid}/storage/{eid}/latest", func(w http.ResponseWriter, r *http.Request) {
		sid := r.PathValue("sid")
		eid := r.PathValue("eid")
		data, err := client.GetStorageLatest(r.Context(), sid, eid)
		writeJSON(w, data, err)
	})

	// ── Health check ──
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "ok"}, nil)
	})

	// ── Dashboard SPA ──
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		sid := os.Getenv("APS_SYS_ID")
		html := strings.ReplaceAll(string(dashboardHTML), "##SID##", sid)
		_, _ = w.Write([]byte(html))
	})

	return logMiddleware(logger, corsMiddleware(mux))
}

// Serve starts the dashboard HTTP server. It blocks until ctx is cancelled.
func Serve(ctx context.Context, addr string, client *api.Client, logger *slog.Logger) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: Handler(client, logger),
	}

	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	logger.Info("dashboard listening", "addr", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("dashboard server: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("dashboard request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
