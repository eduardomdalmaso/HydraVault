package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	httpAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/primary/http"
	yoloExporter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/exporter/yolo"
	fsAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/fs"
	phashAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/phash"
	sqliteAdapter "github.com/eduardomdalmaso/HydraVault/internal/adapters/secondary/sqlite"
	"github.com/eduardomdalmaso/HydraVault/internal/application"
	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

func main() {
	fmt.Println("🏛️ HydraVault — Intelligent Vision Data Engine & Active Learning Vault")
	fmt.Println("🔒 Security Mode: Token-Only Enforcement & SQLite WAL Mode Active")

	// 1. Resolve paths
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working dir: %v", err)
	}

	storageDir := filepath.Join(baseDir, "storage")
	datasetsDir := filepath.Join(baseDir, "datasets")
	dbPath := filepath.Join(storageDir, "vault.db")

	// 2. Initialize Secondary Adapters
	fileStore, err := fsAdapter.NewFileStore(datasetsDir)
	if err != nil {
		log.Fatalf("Failed to initialize file store: %v", err)
	}

	db, err := sqliteAdapter.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite WAL database: %v", err)
	}
	defer db.Close()

	datasetRepo := sqliteAdapter.NewDatasetRepository(db)
	frameRepo := sqliteAdapter.NewFrameRepository(db)
	dedup := phashAdapter.NewDeduplicator()
	exporter := yoloExporter.NewExporter(filepath.Join(datasetsDir, "curated"))

	// 3. Initialize Application Services
	ingestService := application.NewIngestService(frameRepo, datasetRepo, fileStore, dedup)
	datasetService := application.NewDatasetService(datasetRepo, frameRepo, exporter)

	// Seed default dataset if none exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if existing, _ := datasetRepo.List(ctx); len(existing) == 0 {
		_ = datasetService.CreateDataset(ctx, &domain.Dataset{
			DatasetID:   "ds_ppe_safety_v1",
			Name:        "EPI & Segurança do Trabalho",
			Description: "Dataset curado com detecções de EPIs para canteiros de obras e fábricas.",
			Task:        domain.TaskDetect,
			Classes: []domain.ClassMetadata{
				{ID: 0, Name: "helmet"},
				{ID: 1, Name: "vest"},
				{ID: 2, Name: "person"},
				{ID: 3, Name: "boots"},
			},
		})
		log.Println("🌱 Seeded initial dataset template: ds_ppe_safety_v1")
	}
	cancel()

	// 4. Initialize Primary HTTP Adapters
	datasetHandler := httpAdapter.NewDatasetHandler(datasetService)
	inboxHandler := httpAdapter.NewInboxHandler(ingestService, fileStore)
	authMiddleware := httpAdapter.NewAuthMiddleware("") // Accepts valid tokens

	router := httpAdapter.NewRouter(datasetHandler, inboxHandler, authMiddleware)

	server := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start Server with Graceful Shutdown
	go func() {
		log.Printf("🚀 HydraVault Daemon running at http://127.0.0.1:8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down HydraVault gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("✅ HydraVault stopped cleanly.")
}
