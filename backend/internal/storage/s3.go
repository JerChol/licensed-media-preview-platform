package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage uploads dervived artifacts to an S3 bucket instead of writing them to local disk.
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates an S3-backend storage client. AWS credentials are picked up automatically from the environment.
func NewS3Storage(ctx context.Context, bucket string, region string) (*S3Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &S3Storage{client: client, bucket: bucket}, nil
}

// UploadBytes uploads content to the given key (path) in the bucket.
func (s *S3Storage) UploadBytes(ctx context.Context, key string, content []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %w", key, err)
	}
	return nil
}

// UploadFile reads a local file and uploads it to the given key.
func (s *S3Storage) UploadFile(ctx context.Context, localPath string, key string, contentType string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file %s: %w", localPath, err)
	}
	return s.UploadBytes(ctx, key, content, contentType)
}

// DownloadObject retrieves an object's bytes and content type from S3.
func (s *S3Storage) DownloadObject(ctx context.Context, key string) ([]byte, string, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to download %s: %w", key, err)
	}
	defer out.Body.Close()

	content, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read %s:%w", key, err)
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}

	return content, contentType, nil
}
