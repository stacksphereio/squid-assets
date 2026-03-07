package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage interface allows swapping between local and cloud storage
type Storage interface {
	Get(path string) ([]byte, error)
	Exists(path string) bool
}

// LocalStorage serves files from local filesystem
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Get(path string) ([]byte, error) {
	// Security: prevent path traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid path: path traversal detected")
	}

	fullPath := filepath.Join(s.basePath, cleanPath)

	// Security: ensure path is within basePath
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	if !strings.HasPrefix(absPath, absBase) {
		return nil, fmt.Errorf("invalid path: outside base directory")
	}

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return io.ReadAll(file)
}

func (s *LocalStorage) Exists(path string) bool {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return false
	}

	fullPath := filepath.Join(s.basePath, cleanPath)

	// Security check
	absBase, err := filepath.Abs(s.basePath)
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(absPath, absBase) {
		return false
	}

	_, err = os.Stat(fullPath)
	return err == nil
}

// Future: BucketStorage implementation for S3/GCS
// type BucketStorage struct {
//     bucketName string
//     client     interface{}
// }
//
// func NewBucketStorage(bucketName string) (*BucketStorage, error) {
//     // Implementation for cloud storage
//     return nil, nil
// }
//
// func (s *BucketStorage) Get(path string) ([]byte, error) {
//     // Implementation for cloud storage
//     return nil, nil
// }
//
// func (s *BucketStorage) Exists(path string) bool {
//     // Implementation for cloud storage
//     return false
// }
