// Package config provides application configuration management.
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Yandex S3 Configuration
	YandexEndpoint       string
	YandexBucket         string
	YandexAccessKeyID    string
	YandexSecretAccessKey string
	YandexS3Region       string

	// Firebase Configuration
	FirebaseCredentialsPath string

	// Application Configuration
	PresignExpiryUpload   time.Duration
	PresignExpiryDownload time.Duration
	LogLevel              string
	AllowedOrigins        []string
	Port                  string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		YandexEndpoint:        getEnv("YANDEX_ENDPOINT", "storage.yandexcloud.net"),
		YandexBucket:          getEnv("YANDEX_BUCKET", ""),
		YandexAccessKeyID:     getEnv("YANDEX_ACCESS_KEY_ID", ""),
		YandexSecretAccessKey: getEnv("YANDEX_SECRET_ACCESS_KEY", ""),
		YandexS3Region:        getEnv("YANDEX_S3_REGION", "ru-central1"),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "/app/firebase-service-account.json"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		AllowedOrigins:        parseCommaSeparated(getEnv("CORS_ALLOWED_ORIGINS", "*")),
		Port:                  getEnv("PORT", "8080"),
	}

	// Parse durations
	if expiry := getEnv("PRESIGNED_URL_EXPIRY_UPLOAD", "300"); expiry != "" {
		if seconds, err := strconv.Atoi(expiry); err == nil {
			cfg.PresignExpiryUpload = time.Duration(seconds) * time.Second
		}
	}
	if expiry := getEnv("PRESIGNED_URL_EXPIRY_DOWNLOAD", "900"); expiry != "" {
		if seconds, err := strconv.Atoi(expiry); err == nil {
			cfg.PresignExpiryDownload = time.Duration(seconds) * time.Second
		}
	}

	// Validation
	if cfg.YandexBucket == "" {
		return nil, errors.New("YANDEX_BUCKET is required")
	}
	if cfg.YandexAccessKeyID == "" {
		return nil, errors.New("YANDEX_ACCESS_KEY_ID is required")
	}
	if cfg.YandexSecretAccessKey == "" {
		return nil, errors.New("YANDEX_SECRET_ACCESS_KEY is required")
	}
	if cfg.PresignExpiryUpload == 0 {
		cfg.PresignExpiryUpload = 5 * time.Minute
	}
	if cfg.PresignExpiryDownload == 0 {
		cfg.PresignExpiryDownload = 15 * time.Minute
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseCommaSeparated(s string) []string {
	if s == "" || s == "*" {
		return []string{"*"}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}