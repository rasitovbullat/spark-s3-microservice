// Package domain contains business entities and error definitions.
package domain

import (
	"errors"
	"strings"
)

// FileType represents the type of file being uploaded.
type FileType string

const (
	FileTypeAvatar  FileType = "avatar"
	FileTypeMessage FileType = "message"
	FileTypeSticker FileType = "sticker"
)

// String returns the string representation of FileType.
func (ft FileType) String() string {
	return string(ft)
}

// AllowedExtensions defines valid extensions for each file type.
var AllowedExtensions = map[FileType][]string{
	FileTypeAvatar:  {"jpg", "jpeg", "png", "webp", "gif"},
	FileTypeMessage: {"jpg", "jpeg", "png", "webp", "gif", "pdf"},
	FileTypeSticker: {"png", "gif", "webp"},
}

// ContentTypes maps file extensions to MIME types.
var ContentTypes = map[string]string{
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"pdf":  "application/pdf",
}

// Domain errors.
var (
	ErrInvalidFileType    = errors.New("invalid file type")
	ErrInvalidExtension   = errors.New("invalid file extension")
	ErrInvalidFileID      = errors.New("invalid file ID format")
	ErrForbidden          = errors.New("access denied")
	ErrFileNotFound       = errors.New("file not found")
	ErrInternalError      = errors.New("internal server error")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// IsValidExtension checks if the extension is allowed for the file type.
func IsValidExtension(fileType FileType, ext string) bool {
	allowed, ok := AllowedExtensions[fileType]
	if !ok {
		return false
	}

	ext = strings.ToLower(strings.TrimPrefix(ext, "."))

	for _, allowedExt := range allowed {
		if allowedExt == ext {
			return true
		}
	}

	return false
}

// GetContentType returns the MIME type for a given extension.
func GetContentType(ext string) string {
	if ct, ok := ContentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

// ParseFileType converts a string to FileType.
func ParseFileType(s string) (FileType, error) {
	switch strings.ToLower(s) {
	case "avatar":
		return FileTypeAvatar, nil
	case "message":
		return FileTypeMessage, nil
	case "sticker":
		return FileTypeSticker, nil
	default:
		return "", ErrInvalidFileType
	}
}

// ExtractOwnerFromFileID extracts the user ID from a file_id.
// file_id format: {type}_{uid}_{timestamp}_{uuid}.{ext}
// Example: avatar_user123_1700000000_a1b2c3d4.jpg
func ExtractOwnerFromFileID(fileID string) (string, error) {
	parts := strings.Split(fileID, "_")
	// Minimum: type_uid_timestamp_uuid.ext = 4 parts
	if len(parts) < 4 {
		return "", ErrInvalidFileID
	}
	// Owner is the second part (index 1)
	return parts[1], nil
}

// ExtractFileTypeFromFileID extracts the file type from a file_id.
func ExtractFileTypeFromFileID(fileID string) (FileType, error) {
	parts := strings.Split(fileID, "_")
	if len(parts) < 4 {
		return "", ErrInvalidFileID
	}
	return ParseFileType(parts[0])
}

// File represents a file entity with metadata.
type File struct {
	ID     string
	Type   FileType
	UserID string
	Ext    string
}

// BuildS3Key constructs the S3 object key from file metadata.
func (f *File) BuildS3Key() string {
	return f.Type.String() + "/" + f.ID
}

// FileIDToS3Key converts a file_id to S3 object key.
// file_id format: {type}_{uid}_{timestamp}_{uuid}.{ext}
func FileIDToS3Key(fileID string) (string, error) {
	fileType, err := ExtractFileTypeFromFileID(fileID)
	if err != nil {
		return "", err
	}
	return fileType.String() + "/" + fileID, nil
}