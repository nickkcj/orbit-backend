package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type StorageService struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	accountID  string
}

type PresignedUploadResult struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int    `json:"expires_in"`
}

func NewStorageService(accountID, accessKeyID, secretAccessKey, bucketName string) (*StorageService, error) {
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil, fmt.Errorf("R2 credentials not configured")
	}

	// R2 endpoint
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Create custom resolver for R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: r2Endpoint,
		}, nil
	})

	// Load AWS config with R2 credentials
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presigner := s3.NewPresignClient(client)

	return &StorageService{
		client:     client,
		presigner:  presigner,
		bucketName: bucketName,
		accountID:  accountID,
	}, nil
}

// GenerateUploadURL generates a presigned URL for uploading a file
func (s *StorageService) GenerateUploadURL(ctx context.Context, tenantID uuid.UUID, filename string, contentType string) (*PresignedUploadResult, error) {
	// Generate unique file key: tenants/{tenant_id}/videos/{uuid}/{filename}
	fileUUID := uuid.New().String()
	fileKey := fmt.Sprintf("tenants/%s/videos/%s/%s", tenantID.String(), fileUUID, filename)

	// Presign PUT request (15 minutes expiry)
	expiresIn := 15 * time.Minute

	presignedReq, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload URL: %w", err)
	}

	return &PresignedUploadResult{
		UploadURL: presignedReq.URL,
		FileKey:   fileKey,
		ExpiresIn: int(expiresIn.Seconds()),
	}, nil
}

// GenerateImageUploadURL generates a presigned URL for uploading an image (avatars, etc)
func (s *StorageService) GenerateImageUploadURL(ctx context.Context, tenantID uuid.UUID, filename string, contentType string) (*PresignedUploadResult, error) {
	// Generate unique file key: tenants/{tenant_id}/images/{uuid}/{filename}
	fileUUID := uuid.New().String()
	fileKey := fmt.Sprintf("tenants/%s/images/%s/%s", tenantID.String(), fileUUID, filename)

	// Presign PUT request (15 minutes expiry)
	expiresIn := 15 * time.Minute

	presignedReq, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(fileKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload URL: %w", err)
	}

	return &PresignedUploadResult{
		UploadURL: presignedReq.URL,
		FileKey:   fileKey,
		ExpiresIn: int(expiresIn.Seconds()),
	}, nil
}

// GetPublicURL returns the public URL for a file (requires R2.dev subdomain enabled)
func (s *StorageService) GetPublicURL(fileKey string) string {
	return fmt.Sprintf("https://%s.r2.dev/%s", s.bucketName, fileKey)
}

// GenerateDownloadURL generates a presigned URL for downloading/viewing a file
func (s *StorageService) GenerateDownloadURL(ctx context.Context, fileKey string, expiresIn time.Duration) (string, error) {
	presignedReq, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("failed to presign download URL: %w", err)
	}

	return presignedReq.URL, nil
}

// DeleteFile deletes a file from storage
func (s *StorageService) DeleteFile(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
