package domain

import "errors"

var (
	ErrInvalidBBox       = errors.New("invalid bounding box coordinates")
	ErrInvalidPolygon    = errors.New("invalid polygon points")
	ErrInvalidDataset    = errors.New("invalid dataset configuration")
	ErrInvalidFrame      = errors.New("invalid frame metadata")
	ErrDatasetNotFound   = errors.New("dataset not found")
	ErrFrameNotFound     = errors.New("frame not found")
	ErrDuplicateFrame    = errors.New("frame already exists or is a perceptual duplicate")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidPath       = errors.New("invalid file path or path traversal detected")
	ErrExportFailed      = errors.New("failed to export dataset")
)
