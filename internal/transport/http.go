package transport

import (
	"encoding/json"
	"example.com/dicom-gateway/internal/app"
	"example.com/dicom-gateway/internal/audit"
	"example.com/dicom-gateway/internal/deid"
	"example.com/dicom-gateway/internal/dicom"
	"example.com/dicom-gateway/internal/repository"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Server struct {
	App   *app.Service
	Repo  *repository.Repository
	Audit *audit.Log
}

func (s *Server) Handler(logger interface{ Info(string, ...any) }) http.Handler {
	mux := http.NewServeMux()
	registerConsole(mux)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/api/v1/ingest/validate", s.validate)
	mux.HandleFunc("/api/v1/ingest", s.ingest)
	mux.HandleFunc("/api/v1/instances", s.instances)
	mux.HandleFunc("/api/v1/instances/", s.instance)
	mux.HandleFunc("/api/v1/destinations", s.destinations)
	mux.HandleFunc("/api/v1/jobs", s.jobs)
	mux.HandleFunc("/api/v1/audit/export", s.auditExport)
	mux.HandleFunc("/dicomweb/studies", s.ingest)
	return RequestID(mux)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
		return
	}
	d, err := s.App.Validate(r.Context(), http.MaxBytesReader(w, r.Body, s.App.Parser.MaxFileBytes))
	if err != nil {
		writeJSON(w, 400, map[string]string{"code": "DICOM_INVALID", "message": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"dataset": d})
}
func (s *Server) ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
		return
	}
	inst, err := s.App.Ingest(r.Context(), http.MaxBytesReader(w, r.Body, s.App.Parser.MaxFileBytes))
	if err != nil {
		writeJSON(w, 400, map[string]string{"code": "INGEST_FAILED", "message": err.Error()})
		return
	}
	s.App.Route(inst)
	writeJSON(w, 201, map[string]any{"instance": inst})
}
func (s *Server) instances(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		items, _ := s.pageItems(r)
		c := parseCursor(r)
		if c.Offset > len(items) {
			c.Offset = len(items)
		}
		end := c.Offset + c.Limit
		if end > len(items) {
			end = len(items)
		}
		writeJSON(w, 200, map[string]any{"items": items[c.Offset:end], "nextCursor": nextCursor(c.Offset, c.Limit, len(items))})
		return
	}
	writeJSON(w, 405, nil)
}

func (s *Server) pageItems(*http.Request) ([]dicom.Instance, error) {
	return s.Repo.ListInstances(), nil
}
func (s *Server) instance(w http.ResponseWriter, r *http.Request) {
	uid := strings.TrimPrefix(r.URL.Path, "/api/v1/instances/")
	if strings.HasSuffix(uid, "/deidentify") {
		uid = strings.TrimSuffix(uid, "/deidentify")
		inst, ok := s.Repo.GetInstance(uid)
		if !ok {
			writeJSON(w, 404, map[string]string{"code": "NOT_FOUND"})
			return
		}
		var p deid.Policy
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		out, rep, err := s.App.Deidentify(inst, p)
		if err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"instance": out, "report": rep})
		return
	}
	if uid == "" {
		writeJSON(w, 404, nil)
		return
	}
	inst, ok := s.Repo.GetInstance(uid)
	if !ok {
		writeJSON(w, 404, map[string]string{"code": "NOT_FOUND"})
		return
	}
	if r.Method == "GET" {
		writeJSON(w, 200, inst)
		return
	}
	writeJSON(w, 405, nil)
}
func (s *Server) destinations(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		writeJSON(w, 200, map[string]any{"items": s.Repo.Destinations()})
		return
	}
	if r.Method == "POST" {
		var d struct {
			ID, Name, AETitle, Endpoint string
			Enabled                     bool
			MaxConcurrent               int
		}
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			writeJSON(w, 400, map[string]string{"error": err.Error()})
			return
		}
		s.Repo.PutDestination(repositoryDest(d))
		writeJSON(w, 201, d)
		return
	}
	writeJSON(w, 405, nil)
}
func repositoryDest(d struct {
	ID, Name, AETitle, Endpoint string
	Enabled                     bool
	MaxConcurrent               int
}) dicom.Destination {
	return dicom.Destination{ID: d.ID, Name: d.Name, AETitle: d.AETitle, Endpoint: d.Endpoint, Enabled: d.Enabled, MaxConcurrent: d.MaxConcurrent}
}
func (s *Server) jobs(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		writeJSON(w, 200, map[string]any{"items": s.Repo.Jobs()})
		return
	}
	writeJSON(w, 405, nil)
}
func (s *Server) auditExport(w http.ResponseWriter, r *http.Request) {
	if !s.Audit.Verify() {
		writeJSON(w, 500, map[string]string{"error": "audit chain invalid"})
		return
	}
	writeJSON(w, 200, map[string]any{"events": s.Audit.Export()})
}

var _ = fmt.Sprintf
var _ = io.EOF
