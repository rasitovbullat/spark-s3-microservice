// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// PresignRepository defines the interface for S3 presigned URL operations.
type PresignRepository interface {
	GeneratePresignedUploadURL(ctx context.Context, key, contentType string, expiresIn time.Duration) (string, time.Time, error)
	GeneratePresignedDownloadURL(ctx context.Context, key string, expiresIn time.Duration) (string, time.Time, error)
	DeleteObject(ctx context.Context, key string) error
	ObjectExists(ctx context.Context, key string) (bool, error)
}

// S3Repository implements PresignRepository for Yandex S3.
type S3Repository struct {
	presignClient *s3.PresignClient
	awsConfig     aws.Config
	endpoint      string
	bucket        string
	region        string
}

// S3Config holds configuration for S3 repository.
type S3Config struct {
	Endpoint        string
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

// NewS3Repository creates a new S3 repository with Yandex S3 configuration.
func NewS3Repository(ctx context.Context, cfg *S3Config) (*S3Repository, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"", // Session token not required for static keys
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s", cfg.Endpoint))
		o.UsePathStyle = true // Yandex S3 requires path-style addressing
	})

	presignClient := s3.NewPresignClient(client)

	return &S3Repository{
		presignClient: presignClient,
		awsConfig:     awsCfg,
		endpoint:      cfg.Endpoint,
		bucket:        cfg.Bucket,
		region:        cfg.Region,
	}, nil
}

// GeneratePresignedUploadURL creates a presigned URL for uploading an object.
// The URL includes Cache-Control header for immutable caching.
func (r *S3Repository) GeneratePresignedUploadURL(
	ctx context.Context,
	key string,
	contentType string,
	expiresIn time.Duration,
) (string, time.Time, error) {
	if expiresIn == 0 {
		expiresIn = 5 * time.Minute
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	expiresAt := time.Now().Add(expiresIn)

	input := &s3.PutObjectInput{
		Bucket:       aws.String(r.bucket),
		Key:          aws.String(key),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("max-age=31536000, immutable"),
	}

	presignedReq, err := r.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return presignedReq.URL, expiresAt, nil
}

// GeneratePresignedDownloadURL creates a presigned URL for downloading an object.
func (r *S3Repository) GeneratePresignedDownloadURL(
	ctx context.Context,
	key string,
	expiresIn time.Duration,
) (string, time.Time, error) {
	if expiresIn == 0 {
		expiresIn = 15 * time.Minute
	}

	expiresAt := time.Now().Add(expiresIn)

	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}

	presignedReq, err := r.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedReq.URL, expiresAt, nil
}

// DeleteObject removes an object from S3.
func (r *S3Repository) DeleteObject(ctx context.Context, key string) error {
	// Create a new S3 client for the delete operation
	client := s3.NewFromConfig(r.awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s", r.endpoint))
		o.UsePathStyle = true
	})

	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// ObjectExists checks if an object exists in S3 using HeadObject.
func (r *S3Repository) ObjectExists(ctx context.Context, key string) (bool, error) {
	// Create a new S3 client for the head operation
	client := s3.NewFromConfig(r.awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s", r.endpoint))
		o.UsePathStyle = true
	})

	_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}