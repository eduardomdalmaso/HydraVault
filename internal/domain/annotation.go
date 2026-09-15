package domain

import (
	"fmt"
	"math"
)

// BBox represents a normalized YOLO bounding box [0.0, 1.0].
type BBox struct {
	ClassID    int     `json:"class_id"`
	ClassName  string  `json:"class_name"`
	XCenter    float64 `json:"x_center"`
	YCenter    float64 `json:"y_center"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Confidence float64 `json:"confidence,omitempty"`
}

// Validate verifies bounding box mathematical invariants.
func (b BBox) Validate() error {
	if b.ClassID < 0 {
		return fmt.Errorf("%w: class_id cannot be negative (%d)", ErrInvalidBBox, b.ClassID)
	}
	if b.XCenter < 0.0 || b.XCenter > 1.0 || math.IsNaN(b.XCenter) {
		return fmt.Errorf("%w: x_center must be between 0.0 and 1.0 (got %f)", ErrInvalidBBox, b.XCenter)
	}
	if b.YCenter < 0.0 || b.YCenter > 1.0 || math.IsNaN(b.YCenter) {
		return fmt.Errorf("%w: y_center must be between 0.0 and 1.0 (got %f)", ErrInvalidBBox, b.YCenter)
	}
	if b.Width <= 0.0 || b.Width > 1.0 || math.IsNaN(b.Width) {
		return fmt.Errorf("%w: width must be between 0.0 and 1.0 (got %f)", ErrInvalidBBox, b.Width)
	}
	if b.Height <= 0.0 || b.Height > 1.0 || math.IsNaN(b.Height) {
		return fmt.Errorf("%w: height must be between 0.0 and 1.0 (got %f)", ErrInvalidBBox, b.Height)
	}
	return nil
}

// YOLOString formats the bbox as a single line for YOLO txt format.
func (b BBox) YOLOString() string {
	return fmt.Sprintf("%d %.6f %.6f %.6f %.6f", b.ClassID, b.XCenter, b.YCenter, b.Width, b.Height)
}

// Point represents a normalized 2D coordinate [0.0, 1.0].
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Polygon represents a segmented instance mask (e.g. from SAM 2).
type Polygon struct {
	ClassID   int     `json:"class_id"`
	ClassName string  `json:"class_name"`
	Points    []Point `json:"points"`
}

// Validate verifies polygon geometry invariants.
func (p Polygon) Validate() error {
	if p.ClassID < 0 {
		return fmt.Errorf("%w: class_id cannot be negative", ErrInvalidPolygon)
	}
	if len(p.Points) < 3 {
		return fmt.Errorf("%w: polygon requires at least 3 points", ErrInvalidPolygon)
	}
	for _, pt := range p.Points {
		if pt.X < 0.0 || pt.X > 1.0 || pt.Y < 0.0 || pt.Y > 1.0 {
			return fmt.Errorf("%w: point coordinates out of bounds [0, 1]: (%.4f, %.4f)", ErrInvalidPolygon, pt.X, pt.Y)
		}
	}
	return nil
}
