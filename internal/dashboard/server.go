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
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mjrgr/apsystems-mcp-server/internal/api"
)

//go:embed dashboard.html
var page []byte

// Server serves the web dashboard and proxies API requests to the APsystems client.
type Server struct {
	client *api.Client
	logger *slog.Logger
	mux    *http.ServeMux
	sysID  string
}

// New creates a dashboard Server and registers all routes.
func New(client *api.Client, logger *slog.Logger, sysID string) *Server {
	s := &Server{client: client, logger: logger, mux: http.NewServeMux(), sysID: sysID}
	s.mux.HandleFunc("GET /", s.handlePage)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/config", s.handleConfig)
	s.mux.HandleFunc("GET /api/system/{sid}", s.handleSystemDetails)
	s.mux.HandleFunc("GET /api/system/{sid}/summary", s.handleSystemSummary)
	s.mux.HandleFunc("GET /api/system/{sid}/energy", s.handleSystemEnergy)
	s.mux.HandleFunc("GET /api/system/{sid}/inverters", s.handleSystemInverters)
	s.mux.HandleFunc("GET /api/system/{sid}/ecu/{eid}/summary", s.handleECUSummary)
	s.mux.HandleFunc("GET /api/system/{sid}/ecu/{eid}/energy", s.handleECUEnergy)
	s.mux.HandleFunc("GET /api/system/{sid}/storage/{eid}/latest", s.handleStorageLatest)
	return s
}

// Serve starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Serve(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("dashboard listen %s: %w", addr, err)
	}
	s.logger.Info("dashboard listening", "addr", ln.Addr().String())

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handlePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.ReplaceAll(string(page), "##SID##", s.sysID)
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleConfig(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"sid": s.sysID})
}

func (s *Server) handleSystemDetails(w http.ResponseWriter, r *http.Request) {
	data, err := s.client.GetSystemDetails(r.Context(), r.PathValue("sid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSystemSummary(w http.ResponseWriter, r *http.Request) {
	data, err := s.client.GetSystemSummary(r.Context(), r.PathValue("sid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleSystemEnergy(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("energy_level")
	if level == "" {
		level = "daily"
	}
	data, err := s.client.GetSystemEnergy(r.Context(), r.PathValue("sid"), level, r.URL.Query().Get("date_range"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(data))
}

func (s *Server) handleSystemInverters(w http.ResponseWriter, r *http.Request) {
	data, err := s.client.GetSystemInverters(r.Context(), r.PathValue("sid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleECUSummary(w http.ResponseWriter, r *http.Request) {
	data, err := s.client.GetECUSummary(r.Context(), r.PathValue("sid"), r.PathValue("eid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleECUEnergy(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("energy_level")
	if level == "" {
		level = "daily"
	}
	data, err := s.client.GetECUEnergy(r.Context(), r.PathValue("sid"), r.PathValue("eid"), level, r.URL.Query().Get("date_range"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, json.RawMessage(data))
}

func (s *Server) handleStorageLatest(w http.ResponseWriter, r *http.Request) {
	data, err := s.client.GetStorageLatest(r.Context(), r.PathValue("sid"), r.PathValue("eid"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
