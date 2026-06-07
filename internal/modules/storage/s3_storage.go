package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage generates presigned S3 URLs.
type S3Storage struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

// S3Config holds S3 connection settings.
type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  string
	BaseURL   string
}

// NewS3Storage creates an S3-backed object store.
func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if cfg.Bucket == "" || cfg.Region == "" {
		return nil, fmt.Errorf("s3 bucket and region are required")
	}

	loadOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		if endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"); endpoint != "" {
			baseURL = endpoint + "/" + cfg.Bucket
		} else {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
		}
	}

	clientOpts := []func(*s3.Options){}
	if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
		endpoint = strings.TrimRight(endpoint, "/")
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	}

	return &S3Storage{
		client:  s3.NewFromConfig(awsCfg, clientOpts...),
		bucket:  cfg.Bucket,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

// PresignPut returns a presigned PUT URL for an object key.
func (s *S3Storage) PresignPut(ctx context.Context, objectKey, contentType string, expires time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)
	result, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &objectKey,
		ContentType: &contentType,
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

// PublicURL returns the public object URL for a stored key.
func (s *S3Storage) PublicURL(objectKey string) string {
	return s.baseURL + "/" + objectKey
}
