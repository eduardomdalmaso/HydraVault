package domain

import (
	"fmt"
	"strings"
	"time"
)

// FrameStatus represents the lifecycle state of a frame in the Vault.
type FrameStatus string

const (
	FrameStatusRaw        FrameStatus = "RAW"          // Ingested in Bronze Inbox
	FrameStatusAutoLabel  FrameStatus = "AUTO_LABELED" // Pre-annotated by SAM 2
	FrameStatusCurated    FrameStatus = "CURATED"      // Approved by Human/Rule
	FrameStatusRejected   FrameStatus = "REJECTED"     // Discarded (blurry, corrupt)
	FrameStatusExported   FrameStatus = "EXPORTED"     // Included in a Gold dataset
)

// SplitType represents the dataset partitioning for training.
type SplitType string

const (
	SplitTrain SplitType = "train"
	SplitVal   SplitType = "val"
	SplitTest  SplitType = "test"
	SplitNone  SplitType = "none"
)

// Frame represents a computer vision image frame with its telemetry and labels.
type Frame struct {
	FrameID     string      `json:"frame_id"`
	DatasetID   string      `json:"dataset_id"`
	CameraID    string      `json:"camera_id"`
	FilePath    string      `json:"file_path"`
	FileName    string      `json:"file_name"`
	PHash       uint64      `json:"phash"`
	Width       int         `json:"width"`
	Height      int         `json:"height"`
	SizeBytes   int64       `json:"size_bytes"`
	Status      FrameStatus `json:"status"`
	Split       SplitType   `json:"split"`
	BBoxes      []BBox      `json:"bboxes"`
	Polygons    []Polygon   `json:"polygons,omitempty"`
	SourceEvent string      `json:"source_event,omitempty"`
	CapturedAt  time.Time   `json:"captured_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Validate verifies frame invariants.
func (f *Frame) Validate() error {
	if strings.TrimSpace(f.FrameID) == "" {
		return fmt.Errorf("%w: frame_id cannot be empty", ErrInvalidFrame)
	}
	if strings.TrimSpace(f.FilePath) == "" {
		return fmt.Errorf("%w: file_path cannot be empty", ErrInvalidFrame)
	}
	if f.Width <= 0 || f.Height <= 0 {
		return fmt.Errorf("%w: width and height must be positive", ErrInvalidFrame)
	}
	for i, b := range f.BBoxes {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("bbox[%d]: %w", i, err)
		}
	}
	for i, p := range f.Polygons {
		if err := p.Validate(); err != nil {
			return fmt.Errorf("polygon[%d]: %w", i, err)
		}
	}
	return nil
}
