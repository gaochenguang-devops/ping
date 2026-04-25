package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"strings"

	"pingtool/internal/scan"
)

//go:embed templates/index.html
var templateFS embed.FS

type handler struct {
	page     *template.Template
	scanner  *scan.Service
	jobStore *scanJobManager
}

func NewHandler(scanner *scan.Service) (http.Handler, error) {
	page, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		return nil, err
	}

	h := &handler{
		page:     page,
		scanner:  scanner,
		jobStore: newScanJobManager(scanner),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleIndex)
	mux.HandleFunc("/api/scan", h.handleScan)
	mux.HandleFunc("/api/scan-jobs", h.handleScanJobs)
	mux.HandleFunc("/api/scan-jobs/", h.handleScanJobByID)
	mux.HandleFunc("/healthz", h.handleHealth)
	return mux, nil
}

func (h *handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.page.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handler) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	req, err := decodeScanRequest(w, r)
	if err != nil {
		return
	}

	resp, err := h.scanner.Scan(r.Context(), req)
	if err != nil {
		writeScanError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *handler) handleScanJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	req, err := decodeScanRequest(w, r)
	if err != nil {
		return
	}

	job, err := h.jobStore.Create(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

func (h *handler) handleScanJobByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/scan-jobs/")
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(path, "/")
	jobID := parts[0]

	if len(parts) == 1 && r.Method == http.MethodGet {
		job, ok := h.jobStore.Snapshot(jobID)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}
		writeJSON(w, http.StatusOK, job)
		return
	}

	if len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost {
		if ok := h.jobStore.Cancel(jobID); !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
			return
		}

		job, _ := h.jobStore.Snapshot(jobID)
		writeJSON(w, http.StatusOK, job)
		return
	}

	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
}

func (h *handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func decodeScanRequest(w http.ResponseWriter, r *http.Request) (scan.ScanRequest, error) {
	body := http.MaxBytesReader(w, r.Body, 1<<20)
	defer body.Close()

	var req scan.ScanRequest
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求体格式无效"})
		return scan.ScanRequest{}, err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求体只能包含一个 JSON 对象"})
		return scan.ScanRequest{}, errors.New("multiple JSON objects")
	}

	return req, nil
}

func writeScanError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	status := http.StatusInternalServerError
	var validationErr scan.ValidationError
	if errors.As(err, &validationErr) {
		status = http.StatusBadRequest
	}

	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
