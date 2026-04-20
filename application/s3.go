package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func initS3(ctx context.Context, a *Application) error {
	if a.S3 != nil {
		return nil
	}

	cfg := a.s3Config
	if cfg == nil {
		parsed := S3ConfigFromEnv()
		cfg = &parsed
	}

	client, err := newMinioClient(ctx, *cfg)
	if err != nil {
		return err
	}

	a.S3 = client
	bucket := cfg.Bucket
	if bucket != "" {
		a.RegisterReadinessCheck("s3", func(ctx context.Context) error {
			_, err := client.BucketExists(ctx, bucket)
			return err
		})
	}

	return nil
}

func newMinioClient(ctx context.Context, cfg S3Config) (*minio.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	lookup := minio.BucketLookupAuto
	if cfg.ForcePathStyle {
		lookup = minio.BucketLookupPath
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: lookup,
	})
	if err != nil {
		return nil, fmt.Errorf("init s3 client: %w", err)
	}

	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" || !cfg.CheckBucket {
		return client, nil
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check s3 bucket %q: %w", bucket, err)
	}
	if exists {
		return client, nil
	}

	if !cfg.AutoCreateBucket {
		return nil, fmt.Errorf("s3 bucket %q does not exist", bucket)
	}

	if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: cfg.Region}); err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code != "BucketAlreadyOwnedByYou" && resp.Code != "BucketAlreadyExists" {
			return nil, fmt.Errorf("create s3 bucket %q: %w", bucket, err)
		}
	}

	return client, nil
}
