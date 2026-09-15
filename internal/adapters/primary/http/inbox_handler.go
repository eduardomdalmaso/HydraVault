package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/eduardomdalmaso/HydraVault/internal/application"
	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// InboxHandler handles frame uploads and direct inspection.
type InboxHandler struct {
	ingestService *application.IngestService
	fileStore     ports.IFileStore
}

// NewInboxHandler creates a new InboxHandler.
func NewInboxHandler(ingestService *application.IngestService, fileStore ports.IFileStore) *InboxHandler {
	return &InboxHandler{
		ingestService: ingestService,
		fileStore:     fileStore,
	}
}

// HandleUpload receives multipart image uploads and executes active learning ingest.
func (h *InboxHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Max 32MB upload
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'image' form file")
		return
	}
	defer file.Close()

	datasetID := strings.TrimSpace(r.FormValue("dataset_id"))
	if datasetID == "" {
		writeError(w, http.StatusBadRequest, "missing 'dataset_id'")
		return
	}

	cameraID := strings.TrimSpace(r.FormValue("camera_id"))
	if cameraID == "" {
		cameraID = "cam_manual_upload"
	}

	var bboxes []domain.BBox
	bboxesJSON := r.FormValue("bboxes")
	if bboxesJSON != "" {
		_ = json.Unmarshal([]byte(bboxesJSON), &bboxes)
	}

	frame, err := h.ingestService.IngestFrame(r.Context(), application.IngestCommand{
		DatasetID:   datasetID,
		CameraID:    cameraID,
		FileName:    header.Filename,
		ImageReader: file,
		BBoxes:      bboxes,
		SourceEvent: r.FormValue("source_event"),
	})
	if errors.Is(err, domain.ErrDuplicateFrame) {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":    err.Error(),
			"frame_id": frame.FrameID,
			"status":   "duplicate_skipped",
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, frame)
}
