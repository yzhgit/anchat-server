package oss

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"flamingo/pkg/config"

	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ProviderSet = wire.NewSet(NewMinioClient)

// Client MinIO client wrapper
type MinioClient struct {
	client *minio.Client
}

// NewClient creates a new MinIO client
func NewMinioClient(conf config.Minio) (*MinioClient, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKey, conf.SecretKey, ""),
		Secure: conf.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	client := &MinioClient{
		client: minioClient,
	}

	// Ensure all buckets exist
	ctx := context.Background()
	for _, bucket := range conf.Buckets {
		exists, err := minioClient.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("failed to check bucket %s: %w", bucket, err)
		}

		if !exists {
			err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
	}

	return client, nil
}

// PutObject uploads an object
func (c *MinioClient) PutObject(ctx context.Context, bucketName, objectName, filePath string, contentType string) (minio.UploadInfo, error) {
	return c.client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// PutObjectReader uploads an object from reader
func (c *MinioClient) PutObjectReader(ctx context.Context, bucketName, objectName string, reader interface{}, objectSize int64, contentType string) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, bucketName, objectName, reader.(interface {
		Read(p []byte) (n int, err error)
	}), objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

// GetObject gets an object
func (c *MinioClient) GetObject(ctx context.Context, bucketName, objectName string) (*minio.Object, error) {
	return c.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

// RemoveObject deletes an object
func (c *MinioClient) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// StatObject gets object metadata
func (c *MinioClient) StatObject(ctx context.Context, bucketName, objectName string) (minio.ObjectInfo, error) {
	return c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
}

// PresignedGetObject generates a download presigned URL
// Default expiration is 1 hour
func (c *MinioClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	if expires == 0 {
		expires = time.Hour
	}
	return c.client.PresignedGetObject(ctx, bucketName, objectName, expires, nil)
}

// PresignedPutObject generates an upload presigned URL
// Default expiration is 1 hour
func (c *MinioClient) PresignedPutObject(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	if expires == 0 {
		expires = time.Hour
	}
	return c.client.PresignedPutObject(ctx, bucketName, objectName, expires)
}

// GetClient gets the underlying minio.Client
func (c *MinioClient) GetClient() *minio.Client {
	return c.client
}
