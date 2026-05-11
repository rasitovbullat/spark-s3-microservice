package domain

import (
	"testing"
)

func TestParseFileType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected FileType
		wantErr  bool
	}{
		{"valid avatar", "avatar", FileTypeAvatar, false},
		{"valid message", "message", FileTypeMessage, false},
		{"valid sticker", "sticker", FileTypeSticker, false},
		{"uppercase avatar", "Avatar", FileTypeAvatar, false},
		{"uppercase message", "MESSAGE", FileTypeMessage, false},
		{"invalid type", "video", "", true},
		{"empty string", "", "", true},
		{"random string", "random", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFileType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFileType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ParseFileType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidExtension(t *testing.T) {
	tests := []struct {
		name      string
		fileType  FileType
		ext       string
		wantValid bool
	}{
		// Avatar tests
		{"avatar jpg", FileTypeAvatar, "jpg", true},
		{"avatar jpeg", FileTypeAvatar, "jpeg", true},
		{"avatar png", FileTypeAvatar, "png", true},
		{"avatar webp", FileTypeAvatar, "webp", true},
		{"avatar gif", FileTypeAvatar, "gif", true},
		{"avatar pdf", FileTypeAvatar, "pdf", false},
		{"avatar with dot", FileTypeAvatar, ".jpg", true},

		// Message tests
		{"message jpg", FileTypeMessage, "jpg", true},
		{"message pdf", FileTypeMessage, "pdf", true},
		{"message exe", FileTypeMessage, "exe", false},

		// Sticker tests
		{"sticker png", FileTypeSticker, "png", true},
		{"sticker gif", FileTypeSticker, "gif", true},
		{"sticker jpg", FileTypeSticker, "jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidExtension(tt.fileType, tt.ext)
			if result != tt.wantValid {
				t.Errorf("IsValidExtension(%v, %v) = %v, want %v", tt.fileType, tt.ext, result, tt.wantValid)
			}
		})
	}
}

func TestExtractOwnerFromFileID(t *testing.T) {
	tests := []struct {
		name     string
		fileID   string
		expected string
		wantErr  bool
	}{
		{"valid file_id", "avatar_user123_1700000000_a1b2c3d4.jpg", "user123", false},
		{"message file_id", "message_user456_1700000000_b2c3d4e5.png", "user456", false},
		{"sticker file_id", "sticker_user789_1700000000_c3d4e5f6.gif", "user789", false},
		{"too few parts", "avatar_user123", "", true},
		{"empty", "", "", true},
		{"missing timestamp", "avatar_user.png", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractOwnerFromFileID(tt.fileID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractOwnerFromFileID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ExtractOwnerFromFileID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractFileTypeFromFileID(t *testing.T) {
	tests := []struct {
		name     string
		fileID   string
		expected FileType
		wantErr  bool
	}{
		{"avatar file", "avatar_user123_1700000000_a1b2c3d4.jpg", FileTypeAvatar, false},
		{"message file", "message_user456_1700000000_b2c3d4e5.png", FileTypeMessage, false},
		{"sticker file", "sticker_user789_1700000000_c3d4e5f6.gif", FileTypeSticker, false},
		{"invalid type in file_id", "video_user123_1700000000_a1b2c3d4.jpg", "", true},
		{"too few parts", "avatar_user123", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractFileTypeFromFileID(tt.fileID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFileTypeFromFileID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ExtractFileTypeFromFileID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFileIDToS3Key(t *testing.T) {
	tests := []struct {
		name     string
		fileID   string
		expected string
		wantErr  bool
	}{
		{"avatar file", "avatar_user123_1700000000_a1b2c3d4.jpg", "avatar/avatar_user123_1700000000_a1b2c3d4.jpg", false},
		{"message file", "message_user456_1700000000_b2c3d4e5.png", "message/message_user456_1700000000_b2c3d4e5.png", false},
		{"sticker file", "sticker_user789_1700000000_c3d4e5f6.gif", "sticker/sticker_user789_1700000000_c3d4e5f6.gif", false},
		{"invalid file_id", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FileIDToS3Key(tt.fileID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileIDToS3Key() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("FileIDToS3Key() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{"jpg", "image/jpeg"},
		{"jpeg", "image/jpeg"},
		{"png", "image/png"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
		{"pdf", "application/pdf"},
		{"unknown", "application/octet-stream"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := GetContentType(tt.ext)
			if result != tt.expected {
				t.Errorf("GetContentType(%v) = %v, want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestFileTypeString(t *testing.T) {
	tests := []struct {
		fileType FileType
		expected string
	}{
		{FileTypeAvatar, "avatar"},
		{FileTypeMessage, "message"},
		{FileTypeSticker, "sticker"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.fileType.String()
			if result != tt.expected {
				t.Errorf("FileType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}