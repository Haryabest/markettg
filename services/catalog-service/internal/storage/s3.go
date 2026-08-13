package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3 struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewS3(cfg config.S3) (*S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3{client: client, bucket: cfg.Bucket, publicURL: cfg.PublicURL}, nil
}

type PresignResult struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
	PublicURL string `json:"public_url"`
}

func (s *S3) PresignUpload(ctx context.Context, contentType string) (*PresignResult, error) {
	key := fmt.Sprintf("products/%s", uuid.New().String())
	url, err := s.client.PresignedPutObject(ctx, s.bucket, key, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	return &PresignResult{
		UploadURL: url.String(),
		Key:       key,
		PublicURL: fmt.Sprintf("%s/%s", s.publicURL, key),
	}, nil
}

func (s *S3) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	}
	return nil
}
