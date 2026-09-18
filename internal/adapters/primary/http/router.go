package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Router sets up HTTP handlers and middleware.
func NewRouter(
	datasetHandler *DatasetHandler,
	inboxHandler *InboxHandler,
	authMiddleware *AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	// Public Healthcheck
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "online",
			"service": "HydraVault",
			"role":    "Curation Plane & Active Learning Vault",
			"version": "1.0.0",
		})
	})

	// Protected Endpoints
	mux.HandleFunc("/api/v1/datasets", datasetHandler.HandleDatasets)
	mux.HandleFunc("/api/v1/datasets/", datasetHandler.HandleDatasetItem)
	mux.HandleFunc("/api/v1/inbox/upload", inboxHandler.HandleUpload)

	// Wrap mux with AuthMiddleware, Security Headers and CORS
	return WithCORS(authMiddleware.Wrap(securityHeaders(mux)))
}

// WithCORS sets Cross-Origin Resource Sharing headers and handles OPTIONS preflights.
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		next.ServeHTTP(w, r)
	})
}

// ProblemDetails follows RFC 7807 problem detail standard.
type ProblemDetails struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	title := http.StatusText(status)
	problem := ProblemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: strings.TrimSpace(detail),
	}
	writeJSON(w, status, problem)
}
