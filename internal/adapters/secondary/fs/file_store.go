package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/eduardomdalmaso/HydraVault/internal/domain"
)

// FileStore provides safe, sandboxed disk storage for datasets.
type FileStore struct {
	baseDir string
}

// NewFileStore initializes a sandboxed file store with a canonical base directory.
func NewFileStore(baseDir string) (*FileStore, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute base dir: %w", err)
	}
	if err := os.MkdirAll(absBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base dir %s: %w", absBase, err)
	}
	return &FileStore{baseDir: absBase}, nil
}

// resolveSafePath verifies that the requested relative path does not escape the baseDir.
func (s *FileStore) resolveSafePath(relativePath string) (string, error) {
	cleanRel := filepath.Clean(relativePath)
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("%w: path traversal attempt (%s)", domain.ErrInvalidPath, relativePath)
	}
	targetPath := filepath.Join(s.baseDir, cleanRel)
	relCheck, err := filepath.Rel(s.baseDir, targetPath)
	if err != nil || strings.HasPrefix(relCheck, "..") {
		return "", fmt.Errorf("%w: target path escapes sandbox (%s)", domain.ErrInvalidPath, relativePath)
	}
	return targetPath, nil
}

// Save streams content into a sandboxed destination file.
func (s *FileStore) Save(ctx context.Context, relativePath string, r io.Reader) (string, int64, error) {
	targetPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return "", 0, err
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	n, err := io.Copy(file, r)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write data: %w", err)
	}

	return targetPath, n, nil
}

// Open opens a sandboxed file for reading.
func (s *FileStore) Open(ctx context.Context, relativePath string) (io.ReadCloser, error) {
	targetPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return nil, err
	}
	return os.Open(targetPath)
}

// Exists checks if a sandboxed file exists.
func (s *FileStore) Exists(ctx context.Context, relativePath string) bool {
	targetPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return false
	}
	_, err = os.Stat(targetPath)
	return err == nil
}

// Delete removes a sandboxed file.
func (s *FileStore) Delete(ctx context.Context, relativePath string) error {
	targetPath, err := s.resolveSafePath(relativePath)
	if err != nil {
		return err
	}
	return os.Remove(targetPath)
}

// GetAbsolutePath returns the verified absolute path.
func (s *FileStore) GetAbsolutePath(relativePath string) (string, error) {
	return s.resolveSafePath(relativePath)
}
