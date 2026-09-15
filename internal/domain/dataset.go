package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var validDatasetIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

// TaskType represents the vision task.
type TaskType string

const (
	TaskDetect  TaskType = "detect"
	TaskSegment TaskType = "segment"
	TaskPose    TaskType = "pose"
	TaskOBB     TaskType = "obb"
)

// ClassMetadata tracks class mapping and instance counts.
type ClassMetadata struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Dataset represents a curated dataset product ready for training in HydraForge.
type Dataset struct {
	DatasetID    string          `json:"dataset_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Task         TaskType        `json:"task"`
	Classes      []ClassMetadata `json:"classes"`
	TrainImages  int             `json:"train_images"`
	ValImages    int             `json:"val_images"`
	TestImages   int             `json:"test_images"`
	TotalFrames  int             `json:"total_frames"`
	SizeBytes    int64           `json:"size_bytes"`
	YAMLPath     string          `json:"yaml_path,omitempty"`
	ChecksumSHA  string          `json:"checksum_sha,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// Validate verifies dataset configuration invariants.
func (d *Dataset) Validate() error {
	if !validDatasetIDRegex.MatchString(d.DatasetID) {
		return fmt.Errorf("%w: dataset_id must be alphanumeric and between 3-64 chars", ErrInvalidDataset)
	}
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidDataset)
	}
	if len(d.Classes) == 0 {
		return fmt.Errorf("%w: dataset must contain at least 1 class", ErrInvalidDataset)
	}
	seen := make(map[int]bool)
	names := make(map[string]bool)
	for _, c := range d.Classes {
		if c.ID < 0 {
			return fmt.Errorf("%w: class ID cannot be negative (%d)", ErrInvalidDataset, c.ID)
		}
		if seen[c.ID] {
			return fmt.Errorf("%w: duplicate class ID (%d)", ErrInvalidDataset, c.ID)
		}
		seen[c.ID] = true
		cName := strings.TrimSpace(strings.ToLower(c.Name))
		if cName == "" {
			return fmt.Errorf("%w: class name cannot be empty", ErrInvalidDataset)
		}
		if names[cName] {
			return fmt.Errorf("%w: duplicate class name (%s)", ErrInvalidDataset, c.Name)
		}
		names[cName] = true
	}
	return nil
}

// ClassNames returns the ordered slice of class names.
func (d *Dataset) ClassNames() []string {
	res := make([]string, len(d.Classes))
	for i, c := range d.Classes {
		res[i] = c.Name
	}
	return res
}
