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
		var rawList []map[string]any
		if err := json.Unmarshal([]byte(bboxesJSON), &rawList); err == nil {
			for _, item := range rawList {
				bbox := parseFlexibleBBox(item)
				if err := bbox.Validate(); err == nil {
					bboxes = append(bboxes, bbox)
				}
			}
		}
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

func parseFlexibleBBox(m map[string]any) domain.BBox {
	var bbox domain.BBox

	if id, ok := m["class_id"].(float64); ok {
		bbox.ClassID = int(id)
	}
	if name, ok := m["class_name"].(string); ok {
		bbox.ClassName = strings.ToLower(strings.TrimSpace(name))
	} else if lbl, ok := m["label"].(string); ok {
		bbox.ClassName = strings.ToLower(strings.TrimSpace(lbl))
	}

	if conf, ok := m["confidence"].(float64); ok {
		bbox.Confidence = conf
	}

	// 1. Direct x_center, y_center, width, height
	if xc, ok := m["x_center"].(float64); ok {
		bbox.XCenter = xc
		if yc, ok := m["y_center"].(float64); ok {
			bbox.YCenter = yc
		}
		if w, ok := m["width"].(float64); ok {
			bbox.Width = w
		}
		if h, ok := m["height"].(float64); ok {
			bbox.Height = h
		}
	} else if rawBox, ok := m["box"].([]any); ok && len(rawBox) >= 4 {
		// [x, y, w, h] format
		x, _ := rawBox[0].(float64)
		y, _ := rawBox[1].(float64)
		w, _ := rawBox[2].(float64)
		h, _ := rawBox[3].(float64)

		// Convert from 0-100 percentage to 0.0-1.0 if needed
		if x > 1.0 || y > 1.0 || w > 1.0 || h > 1.0 {
			x /= 100.0
			y /= 100.0
			w /= 100.0
			h /= 100.0
		}
		bbox.XCenter = x + (w / 2.0)
		bbox.YCenter = y + (h / 2.0)
		bbox.Width = w
		bbox.Height = h
	}

	// Clamp boundaries safely
	if bbox.Width <= 0.0 {
		bbox.Width = 0.1
	}
	if bbox.Height <= 0.0 {
		bbox.Height = 0.1
	}
	if bbox.XCenter <= 0.0 {
		bbox.XCenter = bbox.Width / 2.0
	}
	if bbox.YCenter <= 0.0 {
		bbox.YCenter = bbox.Height / 2.0
	}

	return bbox
}
