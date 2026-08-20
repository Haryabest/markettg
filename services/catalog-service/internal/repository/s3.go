package repository

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/markettg/markettg/packages/go-shared/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PresignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Prefix      string `json:"prefix"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	Key       string `json:"key"`
	PublicURL string `json:"public_url"`
	ExpiresIn int    `json:"expires_in"`
}

type S3Repository struct {
	client    *minio.Client
	bucket    string
	publicURL string
	expiry    time.Duration
}

func NewS3Repository(cfg config.S3) (*S3Repository, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}
	return &S3Repository{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
		expiry:    15 * time.Minute,
	}, nil
}

func (s *S3Repository) PresignUpload(ctx context.Context, req PresignRequest) (*PresignResponse, error) {
	if req.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	prefix := strings.Trim(req.Prefix, "/")
	if prefix == "" {
		prefix = "uploads"
	}

	ext := path.Ext(req.Filename)
	objectKey := fmt.Sprintf("%s/%s%s", prefix, uuid.New().String(), ext)

	url, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, s.expiry)
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}

	return &PresignResponse{
		UploadURL: url.String(),
		Key:       objectKey,
		PublicURL: fmt.Sprintf("%s/%s", s.publicURL, objectKey),
		ExpiresIn: int(s.expiry.Seconds()),
	}, nil
}

func (s *S3Repository) PublicURL(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.publicURL, strings.TrimLeft(key, "/"))
}

func (s *S3Repository) PutBytes(ctx context.Context, key string, data []byte, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, strings.TrimLeft(key, "/"), bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func ApplyImageURLs(products []Product, s3 *S3Repository) {
	for i := range products {
		ApplyImageURL(&products[i], s3)
	}
}

func ApplyImageURL(p *Product, s3 *S3Repository) {
	if p.ImageKey != nil && *p.ImageKey != "" {
		url := s3.PublicURL(*p.ImageKey)
		p.ImageURL = &url
	}
}

func ApplyBannerURLs(promos []Promotion, s3 *S3Repository) {
	for i := range promos {
		if promos[i].BannerKey != nil && *promos[i].BannerKey != "" {
			url := s3.PublicURL(*promos[i].BannerKey)
			promos[i].BannerURL = &url
		}
	}
}
