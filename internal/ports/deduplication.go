package ports

import (
	"context"
	"image"
	"io"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

// IDeduplicator computes perceptual hashes and measures similarity distances.
type IDeduplicator interface {
	ComputePHash(img image.Image) (uint64, error)
	HammingDistance(hashA, hashB uint64) int
}

// IFileStore provides sandboxed, path-traversal safe storage for image blobs.
type IFileStore interface {
	Save(ctx context.Context, relativePath string, r io.Reader) (string, int64, error)
	Open(ctx context.Context, relativePath string) (io.ReadCloser, error)
	Exists(ctx context.Context, relativePath string) bool
	Delete(ctx context.Context, relativePath string) error
	GetAbsolutePath(relativePath string) (string, error)
}

// ExportResult contains paths and hashes of the exported dataset.
type ExportResult struct {
	DatasetID   string
	YAMLPath    string
	OutputDir   string
	TrainCount  int
	ValCount    int
	TestCount   int
	ChecksumSHA string
}

// IYOLOExporter generates normalized .txt labels and data.yaml for Ultralytics/HydraForge.
type IYOLOExporter interface {
	Export(ctx context.Context, dataset *domain.Dataset, frames []*domain.Frame) (*ExportResult, error)
}
