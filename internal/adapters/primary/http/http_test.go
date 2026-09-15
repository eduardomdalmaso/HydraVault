package http_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	httpAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/primary/http"
	yoloExporter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/exporter/yolo"
	fsAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/fs"
	phashAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/phash"
	sqliteAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/sqlite"
	"github.com/eduardomdalmaso/HydraVault/internal/application"
)

func createTestImage(t *testing.T, w, h int, c color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("failed to encode test jpeg: %v", err)
	}
	return buf.Bytes()
}

func setupTestApp(t *testing.T) (http.Handler, func()) {
	tmpDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test_vault.db")
	db, err := sqliteAdapter.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	fileStore, err := fsAdapter.NewFileStore(filepath.Join(tmpDir, "datasets"))
	if err != nil {
		t.Fatalf("failed to create file store: %v", err)
	}

	datasetRepo := sqliteAdapter.NewDatasetRepository(db)
	frameRepo := sqliteAdapter.NewFrameRepository(db)
	dedup := phashAdapter.NewDeduplicator()
	exporter := yoloExporter.NewExporter(filepath.Join(tmpDir, "datasets", "curated"))

	ingestService := application.NewIngestService(frameRepo, datasetRepo, fileStore, dedup)
	datasetService := application.NewDatasetService(datasetRepo, frameRepo, exporter)

	datasetHandler := httpAdapter.NewDatasetHandler(datasetService)
	inboxHandler := httpAdapter.NewInboxHandler(ingestService, fileStore)
	authMiddleware := httpAdapter.NewAuthMiddleware("test-secret-token")

	handler := httpAdapter.NewRouter(datasetHandler, inboxHandler, authMiddleware)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return handler, cleanup
}

func TestHTTP_AuthMiddleware(t *testing.T) {
	handler, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Health is public
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d", w.Code)
	}

	// 2. /api/v1/datasets without token -> 401 Unauthorized
	req = httptest.NewRequest(http.MethodGet, "/api/v1/datasets", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
	}

	// 3. /api/v1/datasets with invalid token -> 403 Forbidden
	req = httptest.NewRequest(http.MethodGet, "/api/v1/datasets", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", w.Code)
	}

	// 4. /api/v1/datasets with valid Bearer token -> 200 OK
	req = httptest.NewRequest(http.MethodGet, "/api/v1/datasets", nil)
	req.Header.Set("Authorization", "Bearer test-secret-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}
}

func TestHTTP_IngestAndExportFlow(t *testing.T) {
	handler, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Create Dataset via POST
	datasetJSON := `{
		"dataset_id": "ds_traffic_test",
		"name": "Traffic Cam Test",
		"task": "detect",
		"classes": [{"id": 0, "name": "car"}, {"id": 1, "name": "truck"}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/datasets", bytes.NewBufferString(datasetJSON))
	req.Header.Set("Authorization", "Bearer test-secret-token")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Ingest an image
	imgData := createTestImage(t, 100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("dataset_id", "ds_traffic_test")
	_ = writer.WriteField("camera_id", "cam_01")
	part, _ := writer.CreateFormFile("image", "frame_01.jpg")
	_, _ = io.Copy(part, bytes.NewReader(imgData))
	writer.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/v1/inbox/upload", &body)
	req.Header.Set("Authorization", "Bearer test-secret-token")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for ingest, got %d: %s", w.Code, w.Body.String())
	}

	// 3. Ingest exact same image -> Duplicate detected
	var dupBody bytes.Buffer
	writerDup := multipart.NewWriter(&dupBody)
	_ = writerDup.WriteField("dataset_id", "ds_traffic_test")
	_ = writerDup.WriteField("camera_id", "cam_01")
	partDup, _ := writerDup.CreateFormFile("image", "frame_01_dup.jpg")
	_, _ = io.Copy(partDup, bytes.NewReader(imgData))
	writerDup.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/v1/inbox/upload", &dupBody)
	req.Header.Set("Authorization", "Bearer test-secret-token")
	req.Header.Set("Content-Type", writerDup.FormDataContentType())
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for duplicate frame, got %d", w.Code)
	}

	// 4. Export Dataset
	req = httptest.NewRequest(http.MethodPost, "/api/v1/datasets/ds_traffic_test/export", nil)
	req.Header.Set("Authorization", "Bearer test-secret-token")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for export, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFileStore_PathTraversalRejection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fs_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	fs, err := fsAdapter.NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to init file store: %v", err)
	}

	traversalPaths := []string{
		"../../etc/passwd",
		"../test.txt",
		"/etc/shadow",
		"sub/../../secret",
	}

	ctx := context.Background()
	for _, p := range traversalPaths {
		_, _, err := fs.Save(ctx, p, bytes.NewReader([]byte("hacked")))
		if err == nil {
			t.Errorf("expected path traversal error for %s, got nil", p)
		}
	}
}
