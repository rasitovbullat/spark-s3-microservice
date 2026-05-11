// Package main is the entry point for the S3 Storage Microservice.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"spark-s3-microservice/internal/config"
	"spark-s3-microservice/internal/handler"
	"spark-s3-microservice/internal/middleware"
	"spark-s3-microservice/internal/repository"
	"spark-s3-microservice/internal/service"
	"spark-s3-microservice/pkg/firebase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	// Initialize Firebase token verifier
	firebaseVerifier, err := firebase.NewFirebaseVerifier(cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase verifier: %v", err)
	}

	// Initialize S3 repository
	s3Repo, err := repository.NewS3Repository(ctx, &repository.S3Config{
		Endpoint:        cfg.YandexEndpoint,
		Bucket:          cfg.YandexBucket,
		Region:          cfg.YandexS3Region,
		AccessKeyID:     cfg.YandexAccessKeyID,
		SecretAccessKey: cfg.YandexSecretAccessKey,
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 repository: %v", err)
	}

	// Initialize services
	storageService := service.NewStorageService(
		s3Repo,
		cfg.PresignExpiryUpload,
		cfg.PresignExpiryDownload,
	)

	// Initialize handlers
	storageHandler := handler.NewStorageHandler(storageService)
	healthHandler := handler.NewHealthHandler()

	// Setup Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	// Health check (no auth required)
	router.GET("/health", healthHandler.Health)

	// API v1 routes (auth required)
	v1 := router.Group("/api/v1")
	v1.Use(middleware.FirebaseAuth(firebaseVerifier))
	{
		storage := v1.Group("/storage")
		{
			storage.GET("/upload-url", storageHandler.GetUploadURL)
			storage.GET("/download-url", storageHandler.GetDownloadURL)
			storage.DELETE("/file", storageHandler.DeleteFile)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting S3 Storage Microservice on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited properly")
}