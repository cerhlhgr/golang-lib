package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/cerhlhgr/golang-lib/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func initS3(ctx context.Context, a *Application) error {
	client, err := newMinioClient(ctx)
	if err != nil {
		return err
	}

	a.S3 = client

	return nil
}

func newMinioClient(ctx context.Context) (*minio.Client, error) {
	endpoint := config.MustString("S3_ENDPOINT")
	accessKey := config.MustString("S3_ACCESS_KEY")
	secretKey := config.MustString("S3_SECRET_KEY")
	region := config.GetString("S3_REGION", "")
	useSSL := config.GetBool("S3_USE_SSL", false)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("init s3 client: %w", err)
	}

	bucket := strings.TrimSpace(config.GetString("S3_BUCKET", ""))
	if bucket == "" {
		return client, nil
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check s3 bucket %q: %w", bucket, err)
	}
	if exists {
		return client, nil
	}

	if !config.GetBool("S3_AUTO_CREATE_BUCKET", false) {
		return nil, fmt.Errorf("s3 bucket %q does not exist", bucket)
	}

	if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		resp := minio.ToErrorResponse(err)
		if resp.Code != "BucketAlreadyOwnedByYou" && resp.Code != "BucketAlreadyExists" {
			return nil, fmt.Errorf("create s3 bucket %q: %w", bucket, err)
		}
	}

	return client, nil
}
