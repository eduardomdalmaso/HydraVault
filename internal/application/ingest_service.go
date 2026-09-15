package application

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// IngestCommand contains payload for image ingestion.
type IngestCommand struct {
	DatasetID   string
	CameraID    string
	FileName    string
	ImageReader io.Reader
	BBoxes      []domain.BBox
	SourceEvent string
	CapturedAt  time.Time
}

// IngestService coordinates frame ingestion, deduplication and storage.
type IngestService struct {
	frameRepo   ports.IFrameRepository
	datasetRepo ports.IDatasetRepository
	fileStore   ports.IFileStore
	dedup       ports.IDeduplicator
}

// NewIngestService creates a new IngestService instance.
func NewIngestService(
	frameRepo ports.IFrameRepository,
	datasetRepo ports.IDatasetRepository,
	fileStore ports.IFileStore,
	dedup ports.IDeduplicator,
) *IngestService {
	return &IngestService{
		frameRepo:   frameRepo,
		datasetRepo: datasetRepo,
		fileStore:   fileStore,
		dedup:       dedup,
	}
}

// IngestFrame ingests an image, computes pHash, checks deduplication and persists it.
func (s *IngestService) IngestFrame(ctx context.Context, cmd IngestCommand) (*domain.Frame, error) {
	// Verify dataset exists
	if _, err := s.datasetRepo.GetByID(ctx, cmd.DatasetID); err != nil {
		return nil, fmt.Errorf("dataset error: %w", err)
	}

	// Buffer image bytes for decoding and storage
	data, err := io.ReadAll(cmd.ImageReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Compute perceptual hash
	phashVal, err := s.dedup.ComputePHash(img)
	if err != nil {
		return nil, fmt.Errorf("failed to compute phash: %w", err)
	}

	// Deduplication check: distance <= 3 is considered duplicate
	if existing, err := s.frameRepo.FindByPHash(ctx, phashVal, 3); err == nil && existing != nil {
		if existing.CameraID == cmd.CameraID {
			return existing, fmt.Errorf("%w (matches %s with distance <= 3)", domain.ErrDuplicateFrame, existing.FrameID)
		}
	}

	frameID := "frm_" + uuid.New().String()[:12]
	ext := filepath.Ext(cmd.FileName)
	if ext == "" {
		ext = ".jpg"
	}
	savedFileName := frameID + ext
	relPath := filepath.Join("inbox", cmd.DatasetID, savedFileName)

	absPath, sizeBytes, err := s.fileStore.Save(ctx, relPath, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to save image blob: %w", err)
	}

	capturedAt := cmd.CapturedAt
	if capturedAt.IsZero() {
		capturedAt = time.Now().UTC()
	}

	frame := &domain.Frame{
		FrameID:     frameID,
		DatasetID:   cmd.DatasetID,
		CameraID:    cmd.CameraID,
		FilePath:    absPath,
		FileName:    savedFileName,
		PHash:       phashVal,
		Width:       width,
		Height:      height,
		SizeBytes:   sizeBytes,
		Status:      domain.FrameStatusRaw,
		Split:       domain.SplitNone,
		BBoxes:      cmd.BBoxes,
		SourceEvent: cmd.SourceEvent,
		CapturedAt:  capturedAt,
	}

	if err := s.frameRepo.Save(ctx, frame); err != nil {
		_ = s.fileStore.Delete(ctx, relPath)
		return nil, fmt.Errorf("failed to save frame metadata: %w", err)
	}

	return frame, nil
}
