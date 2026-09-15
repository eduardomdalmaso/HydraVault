package domain_test

import (
	"testing"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

func TestBBox_Validation(t *testing.T) {
	tests := []struct {
		name    string
		bbox    domain.BBox
		wantErr bool
	}{
		{
			name: "valid bbox",
			bbox: domain.BBox{
				ClassID: 0,
				XCenter: 0.5,
				YCenter: 0.5,
				Width:   0.2,
				Height:  0.4,
			},
			wantErr: false,
		},
		{
			name: "negative class id",
			bbox: domain.BBox{
				ClassID: -1,
				XCenter: 0.5,
				YCenter: 0.5,
				Width:   0.2,
				Height:  0.4,
			},
			wantErr: true,
		},
		{
			name: "x_center out of bounds (> 1.0)",
			bbox: domain.BBox{
				ClassID: 0,
				XCenter: 1.2,
				YCenter: 0.5,
				Width:   0.2,
				Height:  0.4,
			},
			wantErr: true,
		},
		{
			name: "zero width",
			bbox: domain.BBox{
				ClassID: 0,
				XCenter: 0.5,
				YCenter: 0.5,
				Width:   0.0,
				Height:  0.4,
			},
			wantErr: true,
		},
		{
			name: "negative height",
			bbox: domain.BBox{
				ClassID: 0,
				XCenter: 0.5,
				YCenter: 0.5,
				Width:   0.2,
				Height:  -0.1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bbox.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BBox.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataset_Validation(t *testing.T) {
	d := domain.Dataset{
		DatasetID: "ds_traffic_v1",
		Name:      "Traffic Monitoring",
		Task:      domain.TaskDetect,
		Classes: []domain.ClassMetadata{
			{ID: 0, Name: "car"},
			{ID: 1, Name: "truck"},
		},
	}
	if err := d.Validate(); err != nil {
		t.Fatalf("unexpected error for valid dataset: %v", err)
	}

	// Invalid ID
	d.DatasetID = "invalid/id"
	if err := d.Validate(); err == nil {
		t.Errorf("expected error for invalid dataset ID with slash")
	}

	// Duplicate Class Name
	d.DatasetID = "ds_traffic_v1"
	d.Classes = []domain.ClassMetadata{
		{ID: 0, Name: "car"},
		{ID: 1, Name: "CAR"},
	}
	if err := d.Validate(); err == nil {
		t.Errorf("expected error for duplicate class names")
	}
}
