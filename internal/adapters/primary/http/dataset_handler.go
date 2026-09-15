package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/eduardomdalmaso/HydraVault/internal/application"
	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// DatasetHandler exposes REST endpoints for datasets and curation.
type DatasetHandler struct {
	datasetService *application.DatasetService
}

// NewDatasetHandler creates a new DatasetHandler.
func NewDatasetHandler(datasetService *application.DatasetService) *DatasetHandler {
	return &DatasetHandler{datasetService: datasetService}
}

// HandleDatasets handles GET (list) and POST (create) on /api/v1/datasets.
func (h *DatasetHandler) HandleDatasets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		datasets, err := h.datasetService.ListDatasets(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if datasets == nil {
			datasets = []*domain.Dataset{}
		}
		writeJSON(w, http.StatusOK, datasets)

	case http.MethodPost:
		var d domain.Dataset
		if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json payload")
			return
		}
		if err := h.datasetService.CreateDataset(r.Context(), &d); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, d)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleDatasetItem handles GET, DELETE and Export on /api/v1/datasets/{id}.
func (h *DatasetHandler) HandleDatasetItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/datasets/")
	parts := strings.Split(path, "/")
	datasetID := parts[0]

	if datasetID == "" {
		writeError(w, http.StatusBadRequest, "dataset_id required")
		return
	}

	// Sub-resource: /api/v1/datasets/{id}/export
	if len(parts) >= 2 && parts[1] == "export" && r.Method == http.MethodPost {
		res, err := h.datasetService.ExportDataset(r.Context(), datasetID, 0.2)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, res)
		return
	}

	// Sub-resource: /api/v1/datasets/{id}/frames
	if len(parts) >= 2 && parts[1] == "frames" && r.Method == http.MethodGet {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > 100 {
			limit = 50
		}
		frames, total, err := h.datasetService.ListFrames(r.Context(), ports.FrameFilter{
			DatasetID: datasetID,
			Limit:     limit,
			Offset:    offset,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"dataset_id": datasetID,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"frames":     frames,
		})
		return
	}

	if r.Method == http.MethodGet {
		d, err := h.datasetService.GetDataset(r.Context(), datasetID)
		if errors.Is(err, domain.ErrDatasetNotFound) {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, d)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
