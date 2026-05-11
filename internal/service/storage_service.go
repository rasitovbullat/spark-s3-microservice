// Package service contains business logic implementations.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"spark-s3-microservice/internal/domain"
	"spark-s3-microservice/internal/repository"
)

// StorageService provides business logic for storage operations.
type StorageService struct {
	s3Repo                repository.PresignRepository
	presignExpiryUpload   time.Duration
	presignExpiryDownload time.Duration
}

// NewStorageService creates a new storage service.
func NewStorageService(s3Repo repository.PresignRepository, uploadExpiry, downloadExpiry time.Duration) *StorageService {
	return &StorageService{
		s3Repo:                s3Repo,
		presignExpiryUpload:   uploadExpiry,
		presignExpiryDownload: downloadExpiry,
	}
}

// UploadURLResponse contains the response for upload URL generation.
type UploadURLResponse struct {
	FileID    string    `json:"file_id"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// DownloadURLResponse contains the response for download URL generation.
type DownloadURLResponse struct {
	FileID       string    `json:"file_id"`
	DownloadURL  string    `json:"download_url"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// DeleteResponse contains the response for delete operation.
type DeleteResponse struct {
	FileID  string `json:"file_id"`
	Deleted bool   `json:"deleted"`
}

// GenerateUploadURL generates a unique file_id and presigned upload URL.
func (s *StorageService) GenerateUploadURL(
	ctx context.Context,
	userID string,
	fileType domain.FileType,
	extension string,
) (*UploadURLResponse, error) {
	// Validate extension
	ext := strings.ToLower(strings.TrimPrefix(extension, "."))
	if !domain.IsValidExtension(fileType, ext) {
		return nil, domain.ErrInvalidExtension
	}

	// Generate unique file_id: {type}_{uid}_{timestamp}_{short_uuid}.{ext}
	fileID := s.generateFileID(fileType, userID, ext)

	// Build S3 key
	s3Key := fileType.String() + "/" + fileID

	// Get content type
	contentType := domain.GetContentType(ext)

	// Generate presigned upload URL
	uploadURL, expiresAt, err := s.s3Repo.GeneratePresignedUploadURL(
		ctx,
		s3Key,
		contentType,
		s.presignExpiryUpload,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &UploadURLResponse{
		FileID:    fileID,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}, nil
}

// GenerateDownloadURL generates a presigned download URL for a file.
func (s *StorageService) GenerateDownloadURL(
	ctx context.Context,
	fileID string,
) (*DownloadURLResponse, error) {
	// Determine file type from file_id format and build S3 key
	s3Key, err := domain.FileIDToS3Key(fileID)
	if err != nil {
		return nil, err
	}

	// Check if object exists
	exists, err := s.s3Repo.ObjectExists(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to check object existence: %w", err)
	}
	if !exists {
		return nil, domain.ErrFileNotFound
	}

	// Generate presigned download URL
	downloadURL, expiresAt, err := s.s3Repo.GeneratePresignedDownloadURL(
		ctx,
		s3Key,
		s.presignExpiryDownload,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &DownloadURLResponse{
		FileID:      fileID,
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt,
	}, nil
}

// DeleteFile deletes a file from S3 after verifying ownership.
func (s *StorageService) DeleteFile(
	ctx context.Context,
	userID string,
	fileID string,
) (*DeleteResponse, error) {
	// Validate ownership
	ownerID, err := domain.ExtractOwnerFromFileID(fileID)
	if err != nil {
		return nil, domain.ErrInvalidFileID
	}
	if ownerID != userID {
		return nil, domain.ErrForbidden
	}

	// Build S3 key
	s3Key, err := domain.FileIDToS3Key(fileID)
	if err != nil {
		return nil, err
	}

	// Delete from S3
	if err := s.s3Repo.DeleteObject(ctx, s3Key); err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return &DeleteResponse{
		FileID:  fileID,
		Deleted: true,
	}, nil
}

// generateFileID creates a unique file identifier.
// Format: {type}_{uid}_{timestamp}_{short_uuid}.{ext}
func (s *StorageService) generateFileID(fileType domain.FileType, userID, extension string) string {
	timestamp := time.Now().Unix()
	shortUUID := strings.Split(uuid.New().String(), "-")[0]
	return fmt.Sprintf("%s_%s_%d_%s.%s", fileType, userID, timestamp, shortUUID, extension)
}