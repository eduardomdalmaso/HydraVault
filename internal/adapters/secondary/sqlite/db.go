package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenDB opens a SQLite database in WAL mode and initializes tables.
func OpenDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(1) // Single writer mode for SQLite WAL stability
	db.SetMaxIdleConns(1)

	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS datasets (
		dataset_id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		task TEXT NOT NULL DEFAULT 'detect',
		classes_json TEXT NOT NULL DEFAULT '[]',
		train_images INTEGER NOT NULL DEFAULT 0,
		val_images INTEGER NOT NULL DEFAULT 0,
		test_images INTEGER NOT NULL DEFAULT 0,
		total_frames INTEGER NOT NULL DEFAULT 0,
		size_bytes INTEGER NOT NULL DEFAULT 0,
		yaml_path TEXT NOT NULL DEFAULT '',
		checksum_sha TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS frames (
		frame_id TEXT PRIMARY KEY,
		dataset_id TEXT NOT NULL,
		camera_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		phash INTEGER NOT NULL,
		width INTEGER NOT NULL,
		height INTEGER NOT NULL,
		size_bytes INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'RAW',
		split TEXT NOT NULL DEFAULT 'none',
		bboxes_json TEXT NOT NULL DEFAULT '[]',
		polygons_json TEXT NOT NULL DEFAULT '[]',
		source_event TEXT NOT NULL DEFAULT '',
		captured_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (dataset_id) REFERENCES datasets(dataset_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_frames_dataset ON frames(dataset_id);
	CREATE INDEX IF NOT EXISTS idx_frames_status ON frames(status);
	CREATE INDEX IF NOT EXISTS idx_frames_split ON frames(split);
	CREATE INDEX IF NOT EXISTS idx_frames_camera ON frames(camera_id);
	CREATE INDEX IF NOT EXISTS idx_frames_phash ON frames(phash);
	`
	_, err := db.Exec(schema)
	return err
}
