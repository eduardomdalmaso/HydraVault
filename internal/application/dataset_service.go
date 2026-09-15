package application

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// DatasetService handles dataset lifecycle, class curation, and YOLO export.
type DatasetService struct {
	datasetRepo ports.IDatasetRepository
	frameRepo   ports.IFrameRepository
	exporter    ports.IYOLOExporter
}

// NewDatasetService creates a new DatasetService.
func NewDatasetService(
	datasetRepo ports.IDatasetRepository,
	frameRepo ports.IFrameRepository,
	exporter ports.IYOLOExporter,
) *DatasetService {
	return &DatasetService{
		datasetRepo: datasetRepo,
		frameRepo:   frameRepo,
		exporter:    exporter,
	}
}

// CreateDataset initializes a new dataset template.
func (s *DatasetService) CreateDataset(ctx context.Context, d *domain.Dataset) error {
	if err := d.Validate(); err != nil {
		return err
	}
	return s.datasetRepo.Save(ctx, d)
}

// GetDataset retrieves a dataset by ID.
func (s *DatasetService) GetDataset(ctx context.Context, datasetID string) (*domain.Dataset, error) {
	return s.datasetRepo.GetByID(ctx, datasetID)
}

// ListDatasets returns all registered datasets.
func (s *DatasetService) ListDatasets(ctx context.Context) ([]*domain.Dataset, error) {
	return s.datasetRepo.List(ctx)
}

// ListFrames lists frames for a dataset with pagination.
func (s *DatasetService) ListFrames(ctx context.Context, filter ports.FrameFilter) ([]*domain.Frame, int, error) {
	frames, err := s.frameRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.frameRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return frames, count, nil
}

// ExportDataset splits frames into train/val (e.g. 80/20) and generates data.yaml.
func (s *DatasetService) ExportDataset(ctx context.Context, datasetID string, valRatio float64) (*ports.ExportResult, error) {
	d, err := s.datasetRepo.GetByID(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	frames, err := s.frameRepo.List(ctx, ports.FrameFilter{
		DatasetID: datasetID,
		Limit:     100000,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch frames for export: %w", err)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("%w: dataset has no frames", domain.ErrExportFailed)
	}

	if valRatio <= 0.0 || valRatio >= 1.0 {
		valRatio = 0.2 // Default 20% validation split
	}

	// Deterministic split assignment
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(frames), func(i, j int) { frames[i], frames[j] = frames[j], frames[i] })

	valThreshold := int(float64(len(frames)) * valRatio)
	for i, f := range frames {
		if i < valThreshold {
			f.Split = domain.SplitVal
		} else {
			f.Split = domain.SplitTrain
		}
		_ = s.frameRepo.UpdateStatus(ctx, f.FrameID, domain.FrameStatusExported, f.Split)
	}

	exportRes, err := s.exporter.Export(ctx, d, frames)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrExportFailed, err)
	}

	// Update dataset aggregates
	d.TrainImages = exportRes.TrainCount
	d.ValImages = exportRes.ValCount
	d.TestImages = exportRes.TestCount
	d.TotalFrames = len(frames)
	d.YAMLPath = exportRes.YAMLPath
	d.ChecksumSHA = exportRes.ChecksumSHA
	_ = s.datasetRepo.Save(ctx, d)

	return exportRes, nil
}
