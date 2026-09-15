package nats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eduardomdalmaso/HydraVault/internal/application"
	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

// FrameClaimTicket represents the lightweight claim payload sent over NATS.
type FrameClaimTicket struct {
	DatasetID   string        `json:"dataset_id"`
	CameraID    string        `json:"camera_id"`
	FilePath    string        `json:"file_path"`
	FileName    string        `json:"file_name"`
	BBoxes      []domain.BBox `json:"bboxes,omitempty"`
	SourceEvent string        `json:"source_event,omitempty"`
	CapturedAt  time.Time     `json:"captured_at"`
}

// IngestSubscriber consumes Claim Check events from NATS and ingests them into the Vault.
type IngestSubscriber struct {
	ingestService *application.IngestService
}

// NewIngestSubscriber creates a new IngestSubscriber.
func NewIngestSubscriber(ingestService *application.IngestService) *IngestSubscriber {
	return &IngestSubscriber{ingestService: ingestService}
}

// ProcessClaimTicket reads the referenced local file and ingests it.
func (s *IngestSubscriber) ProcessClaimTicket(ctx context.Context, payload []byte) error {
	var ticket FrameClaimTicket
	if err := json.Unmarshal(payload, &ticket); err != nil {
		return fmt.Errorf("failed to unmarshal claim ticket: %w", err)
	}

	data, err := os.ReadFile(ticket.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read claim file %s: %w", ticket.FilePath, err)
	}

	frame, err := s.ingestService.IngestFrame(ctx, application.IngestCommand{
		DatasetID:   ticket.DatasetID,
		CameraID:    ticket.CameraID,
		FileName:    ticket.FileName,
		ImageReader: bytes.NewReader(data),
		BBoxes:      ticket.BBoxes,
		SourceEvent: ticket.SourceEvent,
		CapturedAt:  ticket.CapturedAt,
	})
	if err != nil {
		return fmt.Errorf("ingest error for frame %s: %w", ticket.FileName, err)
	}

	log.Printf("[HydraVault NATS] Ingested Frame: %s (Dataset: %s, %dx%d)", frame.FrameID, frame.DatasetID, frame.Width, frame.Height)
	return nil
}
