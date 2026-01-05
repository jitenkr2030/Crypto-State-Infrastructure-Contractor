package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"forensic-tools/internal/core/ports"
)

// LocalStorage implements ports.StorageBackend using local filesystem
type LocalStorage struct {
	basePath string
	logger   ports.Logger
}

// NewLocalStorage creates a new LocalStorage
func NewLocalStorage(basePath string, logger ports.Logger) (*LocalStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// getFilePath returns the full path for an evidence file
func (s *LocalStorage) getFilePath(evidenceID string) string {
	return filepath.Join(s.basePath, evidenceID[:2], evidenceID[2:4], evidenceID)
}

// StoreEvidence stores evidence file locally
func (s *LocalStorage) StoreEvidence(ctx context.Context, evidenceID string, file io.Reader, metadata *ports.FileMetadata) error {
	filePath := s.getFilePath(evidenceID)

	// Create directory structure
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.logger.Error("Failed to create directory", "error", err, "path", dir)
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	outFile, err := os.Create(filePath)
	if err != nil {
		s.logger.Error("Failed to create file", "error", err, "path", filePath)
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy file content and calculate hash
	hash := sha256.New()
	writer := io.MultiWriter(outFile, hash)

	written, err := io.Copy(writer, file)
	if err != nil {
		s.logger.Error("Failed to write file", "error", err, "path", filePath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Create metadata file
	metaPath := filePath + ".meta"
	metaFile, err := os.Create(metaPath)
	if err != nil {
		s.logger.Error("Failed to create metadata file", "error", err, "path", metaPath)
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metaFile.Close()

	// Write metadata
	metaContent := fmt.Sprintf("filename=%s\ncontent_type=%s\nsize=%d\nhash=%s\ncreated_at=%s\n",
		metadata.Filename,
		metadata.ContentType,
		metadata.Size,
		metadata.Hash,
		metadata.CreatedAt.Format(time.RFC3339),
	)
	if _, err := metaFile.WriteString(metaContent); err != nil {
		s.logger.Error("Failed to write metadata", "error", err, "path", metaPath)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	s.logger.Info("Evidence stored", "id", evidenceID, "path", filePath, "size", written)
	return nil
}

// RetrieveEvidence retrieves evidence file
func (s *LocalStorage) RetrieveEvidence(ctx context.Context, evidenceID string) (io.ReadCloser, *ports.FileMetadata, error) {
	filePath := s.getFilePath(evidenceID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("evidence file not found: %s", evidenceID)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Error("Failed to open file", "error", err, "path", filePath)
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Read metadata
	metadata, err := s.readMetadata(evidenceID)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	return file, metadata, nil
}

// DeleteEvidence deletes evidence file
func (s *LocalStorage) DeleteEvidence(ctx context.Context, evidenceID string) error {
	filePath := s.getFilePath(evidenceID)
	metaPath := filePath + ".meta"

	// Delete metadata file if exists
	if _, err := os.Stat(metaPath); err == nil {
		if err := os.Remove(metaPath); err != nil {
			s.logger.Warn("Failed to delete metadata file", "error", err, "path", metaPath)
		}
	}

	// Delete evidence file
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			s.logger.Error("Failed to delete file", "error", err, "path", filePath)
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	s.logger.Info("Evidence deleted", "id", evidenceID, "path", filePath)
	return nil
}

// EvidenceExists checks if evidence exists
func (s *LocalStorage) EvidenceExists(ctx context.Context, evidenceID string) (bool, error) {
	filePath := s.getFilePath(evidenceID)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetStorageStats returns storage statistics
func (s *LocalStorage) GetStorageStats(ctx context.Context) (*ports.StorageStats, error) {
	var totalSize int64
	var fileCount int64

	err := filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() && !info.IsDir() && filepath.Ext(path) != ".meta" {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	if err != nil {
		s.logger.Error("Failed to get storage stats", "error", err)
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	return &ports.StorageStats{
		TotalSize:   totalSize,
		FileCount:   fileCount,
		LastUpdated: time.Now().UTC(),
	}, nil
}

// readMetadata reads metadata from file
func (s *LocalStorage) readMetadata(evidenceID string) (*ports.FileMetadata, error) {
	filePath := s.getFilePath(evidenceID)
	metaPath := filePath + ".meta"

	metaFile, err := os.Open(metaPath)
	if err != nil {
		s.logger.Error("Failed to open metadata file", "error", err, "path", metaPath)
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer metaFile.Close()

	metadata := &ports.FileMetadata{}

	// Parse metadata
	content := make([]byte, 1024)
	n, _ := metaFile.Read(content)

	lines := string(content[:n])
	for _, line := range []string{
		"filename=",
		"content_type=",
		"size=",
		"hash=",
		"created_at=",
	} {
		if idx := findString(lines, line); idx != -1 {
			endIdx := findString(lines[idx+len(line):], "\n")
			if endIdx == -1 {
				endIdx = len(lines)
			}
			value := lines[idx+len(line) : idx+len(line)+endIdx]

			switch line {
			case "filename=":
				metadata.Filename = value
			case "content_type=":
				metadata.ContentType = value
			case "size=":
				fmt.Sscanf(value, "%d", &metadata.Size)
			case "hash=":
				metadata.Hash = value
			case "created_at=":
				metadata.CreatedAt, _ = time.Parse(time.RFC3339, value)
			}
		}
	}

	return metadata, nil
}

// findString finds a substring in a string
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// S3Storage implements ports.StorageBackend using AWS S3
type S3Storage struct {
	bucket    string
	prefix    string
	s3Client  interface{} // Would be *s3.Client in real implementation
	logger    ports.Logger
}

// NewS3Storage creates a new S3Storage
func NewS3Storage(bucket, prefix string, s3Client interface{}, logger ports.Logger) *S3Storage {
	return &S3Storage{
		bucket:   bucket,
		prefix:   prefix,
		s3Client: s3Client,
		logger:   logger,
	}
}

// StoreEvidence stores evidence to S3
func (s *S3Storage) StoreEvidence(ctx context.Context, evidenceID string, file io.Reader, metadata *ports.FileMetadata) error {
	// In a real implementation, this would upload to S3
	s.logger.Info("S3 storage not implemented, using local storage")
	return fmt.Errorf("S3 storage not implemented")
}

// RetrieveEvidence retrieves evidence from S3
func (s *S3Storage) RetrieveEvidence(ctx context.Context, evidenceID string) (io.ReadCloser, *ports.FileMetadata, error) {
	return nil, nil, fmt.Errorf("S3 storage not implemented")
}

// DeleteEvidence deletes evidence from S3
func (s *S3Storage) DeleteEvidence(ctx context.Context, evidenceID string) error {
	return fmt.Errorf("S3 storage not implemented")
}

// EvidenceExists checks if evidence exists in S3
func (s *S3Storage) EvidenceExists(ctx context.Context, evidenceID string) (bool, error) {
	return false, fmt.Errorf("S3 storage not implemented")
}

// GetStorageStats returns storage statistics from S3
func (s *S3Storage) GetStorageStats(ctx context.Context) (*ports.StorageStats, error) {
	return nil, fmt.Errorf("S3 storage not implemented")
}

// CalculateFileHash calculates SHA256 hash of a file
func CalculateFileHash(file io.Reader) (string, int64, error) {
	hash := sha256.New()
	written, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), written, nil
}
