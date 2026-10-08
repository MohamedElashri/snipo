package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Config holds S3 storage configuration
type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	UseSSL          bool
}

// S3Storage provides S3-compatible object storage operations
type S3Storage struct {
	client *s3.Client
	bucket string
}

// ObjectInfo represents information about an S3 object
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// NewS3Storage creates a new S3 storage client
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Build endpoint URL
	scheme := "https"
	if !cfg.UseSSL {
		scheme = "http"
	}
	endpointURL := fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with service-specific endpoint and path-style addressing
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
		o.UsePathStyle = true
	})

	// Ensure bucket exists
	ctx := context.Background()
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		// Try to create the bucket if it doesn't exist
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			_, createErr := client.CreateBucket(ctx, &s3.CreateBucketInput{
				Bucket: aws.String(cfg.Bucket),
			})
			if createErr != nil {
				return nil, fmt.Errorf("failed to create bucket: %w", createErr)
			}
		} else {
			return nil, fmt.Errorf("failed to check bucket: %w", err)
		}
	}

	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

// Upload uploads content to S3
func (s *S3Storage) Upload(ctx context.Context, key string, content []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	return err
}

// Download retrieves content from S3
func (s *S3Storage) Download(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer func() {
		if err := result.Body.Close(); err != nil {
			slog.Error("failed to close S3 object body", "error", err)
		}
	}()

	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return content, nil
}

// Delete removes an object from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// List returns objects with given prefix
func (s *S3Storage) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			objects = append(objects, ObjectInfo{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
			})
		}
	}

	return objects, nil
}
