package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"strings"
	"time"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
	"github.com/eduardomdalmaso/HydraVault/internal/ports"
)

// FrameRepository implements ports.IFrameRepository using SQLite WAL.
type FrameRepository struct {
	db *sql.DB
}

// NewFrameRepository creates a new FrameRepository.
func NewFrameRepository(db *sql.DB) *FrameRepository {
	return &FrameRepository{db: db}
}

var _ ports.IFrameRepository = (*FrameRepository)(nil)

// Save creates or updates a frame entity.
func (r *FrameRepository) Save(ctx context.Context, f *domain.Frame) error {
	if err := f.Validate(); err != nil {
		return err
	}

	bboxesJSON, err := json.Marshal(f.BBoxes)
	if err != nil {
		return fmt.Errorf("failed to marshal bboxes: %w", err)
	}

	polygonsJSON, err := json.Marshal(f.Polygons)
	if err != nil {
		return fmt.Errorf("failed to marshal polygons: %w", err)
	}

	now := time.Now().UTC()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now

	query := `
	INSERT INTO frames (
		frame_id, dataset_id, camera_id, file_path, file_name,
		phash, width, height, size_bytes, status, split,
		bboxes_json, polygons_json, source_event, captured_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(frame_id) DO UPDATE SET
		dataset_id = excluded.dataset_id,
		camera_id = excluded.camera_id,
		file_path = excluded.file_path,
		file_name = excluded.file_name,
		phash = excluded.phash,
		width = excluded.width,
		height = excluded.height,
		size_bytes = excluded.size_bytes,
		status = excluded.status,
		split = excluded.split,
		bboxes_json = excluded.bboxes_json,
		polygons_json = excluded.polygons_json,
		source_event = excluded.source_event,
		updated_at = excluded.updated_at;
	`

	_, err = r.db.ExecContext(ctx, query,
		f.FrameID, f.DatasetID, f.CameraID, f.FilePath, f.FileName,
		int64(f.PHash), f.Width, f.Height, f.SizeBytes, string(f.Status), string(f.Split),
		string(bboxesJSON), string(polygonsJSON), f.SourceEvent, f.CapturedAt, f.CreatedAt, f.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save frame: %w", err)
	}
	return nil
}

// GetByID retrieves a single frame by ID.
func (r *FrameRepository) GetByID(ctx context.Context, frameID string) (*domain.Frame, error) {
	query := `
	SELECT frame_id, dataset_id, camera_id, file_path, file_name,
	       phash, width, height, size_bytes, status, split,
	       bboxes_json, polygons_json, source_event, captured_at, created_at, updated_at
	FROM frames WHERE frame_id = ?;
	`
	row := r.db.QueryRowContext(ctx, query, frameID)

	var f domain.Frame
	var phashInt int64
	var statusStr, splitStr, bboxesJSON, polygonsJSON string

	err := row.Scan(
		&f.FrameID, &f.DatasetID, &f.CameraID, &f.FilePath, &f.FileName,
		&phashInt, &f.Width, &f.Height, &f.SizeBytes, &statusStr, &splitStr,
		&bboxesJSON, &polygonsJSON, &f.SourceEvent, &f.CapturedAt, &f.CreatedAt, &f.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrFrameNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query frame: %w", err)
	}

	f.PHash = uint64(phashInt)
	f.Status = domain.FrameStatus(statusStr)
	f.Split = domain.SplitType(splitStr)
	_ = json.Unmarshal([]byte(bboxesJSON), &f.BBoxes)
	_ = json.Unmarshal([]byte(polygonsJSON), &f.Polygons)

	return &f, nil
}

// List queries frames matching a filter.
func (r *FrameRepository) List(ctx context.Context, filter ports.FrameFilter) ([]*domain.Frame, error) {
	var sb strings.Builder
	sb.WriteString(`
	SELECT frame_id, dataset_id, camera_id, file_path, file_name,
	       phash, width, height, size_bytes, status, split,
	       bboxes_json, polygons_json, source_event, captured_at, created_at, updated_at
	FROM frames WHERE 1=1`)

	var args []any
	if filter.DatasetID != "" {
		sb.WriteString(" AND dataset_id = ?")
		args = append(args, filter.DatasetID)
	}
	if filter.CameraID != "" {
		sb.WriteString(" AND camera_id = ?")
		args = append(args, filter.CameraID)
	}
	if filter.Status != "" {
		sb.WriteString(" AND status = ?")
		args = append(args, string(filter.Status))
	}
	if filter.Split != "" {
		sb.WriteString(" AND split = ?")
		args = append(args, string(filter.Split))
	}

	sb.WriteString(" ORDER BY captured_at DESC")
	if filter.Limit > 0 {
		sb.WriteString(" LIMIT ?")
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			sb.WriteString(" OFFSET ?")
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list frames: %w", err)
	}
	defer rows.Close()

	var result []*domain.Frame
	for rows.Next() {
		var f domain.Frame
		var phashInt int64
		var statusStr, splitStr, bboxesJSON, polygonsJSON string

		err := rows.Scan(
			&f.FrameID, &f.DatasetID, &f.CameraID, &f.FilePath, &f.FileName,
			&phashInt, &f.Width, &f.Height, &f.SizeBytes, &statusStr, &splitStr,
			&bboxesJSON, &polygonsJSON, &f.SourceEvent, &f.CapturedAt, &f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan frame: %w", err)
		}
		f.PHash = uint64(phashInt)
		f.Status = domain.FrameStatus(statusStr)
		f.Split = domain.SplitType(splitStr)
		_ = json.Unmarshal([]byte(bboxesJSON), &f.BBoxes)
		_ = json.Unmarshal([]byte(polygonsJSON), &f.Polygons)

		result = append(result, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating frames: %w", err)
	}
	return result, nil
}

// Count returns total number of matching frames.
func (r *FrameRepository) Count(ctx context.Context, filter ports.FrameFilter) (int, error) {
	var sb strings.Builder
	sb.WriteString("SELECT COUNT(*) FROM frames WHERE 1=1")
	var args []any
	if filter.DatasetID != "" {
		sb.WriteString(" AND dataset_id = ?")
		args = append(args, filter.DatasetID)
	}
	if filter.Status != "" {
		sb.WriteString(" AND status = ?")
		args = append(args, string(filter.Status))
	}
	var count int
	err := r.db.QueryRowContext(ctx, sb.String(), args...).Scan(&count)
	return count, err
}

// FindByPHash searches for frames within maxHammingDistance.
func (r *FrameRepository) FindByPHash(ctx context.Context, phash uint64, maxDistance int) (*domain.Frame, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT frame_id, phash FROM frames ORDER BY captured_at DESC LIMIT 500")
	if err != nil {
		return nil, err
	}

	var matchedID string
	for rows.Next() {
		var id string
		var pInt int64
		if err := rows.Scan(&id, &pInt); err == nil {
			dist := bits.OnesCount64(phash ^ uint64(pInt))
			if dist <= maxDistance {
				matchedID = id
				break
			}
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("error searching phash: %w", err)
	}
	_ = rows.Close()

	if matchedID != "" {
		return r.GetByID(ctx, matchedID)
	}
	return nil, domain.ErrFrameNotFound
}

// UpdateStatus updates the frame status and split.
func (r *FrameRepository) UpdateStatus(ctx context.Context, frameID string, status domain.FrameStatus, split domain.SplitType) error {
	query := `UPDATE frames SET status = ?, split = ?, updated_at = ? WHERE frame_id = ?;`
	res, err := r.db.ExecContext(ctx, query, string(status), string(split), time.Now().UTC(), frameID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrFrameNotFound
	}
	return nil
}

// Delete removes a frame by ID.
func (r *FrameRepository) Delete(ctx context.Context, frameID string) error {
	query := `DELETE FROM frames WHERE frame_id = ?;`
	_, err := r.db.ExecContext(ctx, query, frameID)
	return err
}
