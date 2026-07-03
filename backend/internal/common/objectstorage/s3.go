package objectstorage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	client *s3.Client
	bucket string
	logger *slog.Logger
}

func NewS3Store(cfg Config, bucket string, logger *slog.Logger) (*S3Store, error) {
	if ok, err := cfg.Validate(); !ok {
		return nil, err
	}

	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return nil, fmt.Errorf("%w: missing bucket", ErrInvalidConfig)
	}

	awsCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &S3Store{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

func (s *S3Store) Check(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf("check object storage bucket %q: %w", s.bucket, err)
	}

	return nil
}

func (s *S3Store) Put(ctx context.Context, input PutObjectInput) (ObjectInfo, error) {
	key := strings.TrimSpace(input.Key)
	if key == "" {
		return ObjectInfo{}, fmt.Errorf("%w: missing object key", ErrInvalidConfig)
	}

	out, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          input.Body,
		ContentLength: aws.Int64(input.Size),
		ContentType:   aws.String(input.ContentType),
	})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("put object %q: %w", key, err)
	}

	return ObjectInfo{
		Key:         key,
		ETag:        strings.Trim(aws.ToString(out.ETag), "\""),
		Size:        input.Size,
		ContentType: input.ContentType,
	}, nil
}

func (s *S3Store) Get(ctx context.Context, key string) (*Object, error) {
	if key == "" {
		return nil, fmt.Errorf("%w: missing object key", ErrInvalidConfig)
	}

	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}

	return &Object{
		Body: out.Body,
		Info: ObjectInfo{
			Key:         key,
			ETag:        strings.Trim(aws.ToString(out.ETag), "\""),
			Size:        aws.ToInt64(out.ContentLength),
			ContentType: aws.ToString(out.ContentType),
		},
	}, nil
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("%w: missing object key", ErrInvalidConfig)
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}

	return nil
}

func (s *S3Store) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("%w: missing object key", ErrInvalidConfig)
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, fmt.Errorf("check object %q: %w", key, err)
	}

	return true, nil
}
