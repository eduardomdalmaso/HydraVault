package ports

import (
	"context"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

// IDatasetRepository defines persistence operations for datasets.
type IDatasetRepository interface {
	Save(ctx context.Context, dataset *domain.Dataset) error
	GetByID(ctx context.Context, datasetID string) (*domain.Dataset, error)
	List(ctx context.Context) ([]*domain.Dataset, error)
	Delete(ctx context.Context, datasetID string) error
}

// FrameFilter provides criteria for listing frames.
type FrameFilter struct {
	DatasetID string
	CameraID  string
	Status    domain.FrameStatus
	Split     domain.SplitType
	Limit     int
	Offset    int
}

// IFrameRepository defines persistence operations for frames.
type IFrameRepository interface {
	Save(ctx context.Context, frame *domain.Frame) error
	GetByID(ctx context.Context, frameID string) (*domain.Frame, error)
	List(ctx context.Context, filter FrameFilter) ([]*domain.Frame, error)
	Count(ctx context.Context, filter FrameFilter) (int, error)
	FindByPHash(ctx context.Context, phash uint64, maxHammingDistance int) (*domain.Frame, error)
	UpdateStatus(ctx context.Context, frameID string, status domain.FrameStatus, split domain.SplitType) error
	Delete(ctx context.Context, frameID string) error
}
