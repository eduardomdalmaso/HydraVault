package yolo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// Exporter formats and freezes datasets into YOLO-compliant structure with data.yaml.
type Exporter struct {
	baseOutputDir string
}

// NewExporter creates a new YOLO Exporter.
func NewExporter(baseOutputDir string) *Exporter {
	return &Exporter{baseOutputDir: baseOutputDir}
}

var _ ports.IYOLOExporter = (*Exporter)(nil)

// Export packages frames and labels into images/labels splits and generates data.yaml.
func (e *Exporter) Export(ctx context.Context, d *domain.Dataset, frames []*domain.Frame) (*ports.ExportResult, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	datasetDir := filepath.Join(e.baseOutputDir, d.DatasetID)
	imagesTrain := filepath.Join(datasetDir, "images", "train")
	imagesVal := filepath.Join(datasetDir, "images", "val")
	labelsTrain := filepath.Join(datasetDir, "labels", "train")
	labelsVal := filepath.Join(datasetDir, "labels", "val")

	for _, dir := range []string{imagesTrain, imagesVal, labelsTrain, labelsVal} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	var trainCount, valCount, testCount int

	for _, f := range frames {
		split := f.Split
		if split == domain.SplitNone || split == "" {
			split = domain.SplitTrain
		}

		subDir := "train"
		switch split {
		case domain.SplitVal:
			subDir = "val"
			valCount++
		case domain.SplitTest:
			subDir = "val" // Ultralytics accepts val/test
			testCount++
		default:
			trainCount++
		}

		// Copy/Link image file
		targetImgPath := filepath.Join(datasetDir, "images", subDir, f.FileName)
		if err := copyFile(f.FilePath, targetImgPath); err != nil {
			return nil, fmt.Errorf("failed to copy frame %s: %w", f.FrameID, err)
		}

		// Generate label .txt
		labelName := strings.TrimSuffix(f.FileName, filepath.Ext(f.FileName)) + ".txt"
		targetLabelPath := filepath.Join(datasetDir, "labels", subDir, labelName)

		var labelContent strings.Builder
		for _, b := range f.BBoxes {
			labelContent.WriteString(b.YOLOString())
			labelContent.WriteByte('\n')
		}
		if err := os.WriteFile(targetLabelPath, []byte(labelContent.String()), 0644); err != nil {
			return nil, fmt.Errorf("failed to write label file %s: %w", targetLabelPath, err)
		}
	}

	// Generate data.yaml
	yamlPath := filepath.Join(datasetDir, "data.yaml")
	var yamlBuilder strings.Builder
	yamlBuilder.WriteString(fmt.Sprintf("# HydraVault Curated Dataset: %s\n", d.Name))
	yamlBuilder.WriteString(fmt.Sprintf("path: %s\n", datasetDir))
	yamlBuilder.WriteString("train: images/train\n")
	yamlBuilder.WriteString("val: images/val\n")
	yamlBuilder.WriteString("\nnames:\n")
	for _, c := range d.Classes {
		yamlBuilder.WriteString(fmt.Sprintf("  %d: %s\n", c.ID, c.Name))
	}

	yamlBytes := []byte(yamlBuilder.String())
	if err := os.WriteFile(yamlPath, yamlBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write data.yaml: %w", err)
	}

	// Calculate SHA256 checksum
	hasher := sha256.New()
	hasher.Write(yamlBytes)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	return &ports.ExportResult{
		DatasetID:   d.DatasetID,
		YAMLPath:    yamlPath,
		OutputDir:   datasetDir,
		TrainCount:  trainCount,
		ValCount:    valCount,
		TestCount:   testCount,
		ChecksumSHA: checksum,
	}, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
