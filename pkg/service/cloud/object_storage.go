package cloud

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type objectStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	publicBaseURL string
}

type noopObjectStorage struct{}

func NewObjectStorageService(cfg config.Config) (CloudService, error) {
	endpoint := firstNonEmpty(cfg.S3Endpoint)
	region := firstNonEmpty(cfg.S3Region, cfg.AwsRegion)
	accessKey := firstNonEmpty(cfg.S3AccessKey, cfg.AwsAccessKeyID)
	secretKey := firstNonEmpty(cfg.S3SecretKey, cfg.AwsSecretKey)
	bucket := firstNonEmpty(cfg.S3Bucket, cfg.AwsBucketName)

	if bucket == "" || accessKey == "" || secretKey == "" {
		return noopObjectStorage{}, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load object storage config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	publicBase := strings.TrimRight(cfg.S3PublicBaseURL, "/")
	if publicBase == "" {
		if endpoint != "" {
			publicBase = strings.TrimRight(endpoint, "/") + "/" + bucket
		} else {
			publicBase = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)
		}
	}

	return &objectStorage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
		publicBaseURL: publicBase,
	}, nil
}

func (noopObjectStorage) SaveFile(ctx context.Context, fh *multipart.FileHeader, opts SaveOptions) (string, error) {
	return "", fmt.Errorf("object storage unavailable: S3 configuration is missing")
}

func (noopObjectStorage) SaveBytes(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
	return "", fmt.Errorf("object storage unavailable: S3 configuration is missing")
}

func (noopObjectStorage) PublicURL(objectKey string) string {
	return objectKey
}

func (noopObjectStorage) PresignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	return "", fmt.Errorf("object storage unavailable: S3 configuration is missing")
}

func (noopObjectStorage) DeleteObject(ctx context.Context, objectKey string) error {
	return nil
}

func (s *objectStorage) SaveFile(ctx context.Context, fh *multipart.FileHeader, opts SaveOptions) (string, error) {
	file, err := fh.Open()
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to open file")
	}
	defer file.Close()

	key := buildKey(opts, fh.Filename)
	contentType := opts.ContentType
	if contentType == "" {
		contentType = fh.Header.Get("Content-Type")
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	}
	if opts.Visibility == VisibilityPublic {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		log.Printf("cloud SaveFile PutObject failed: bucket=%s key=%s contentType=%s visibility=%d filename=%s err=%v", s.bucket, key, contentType, opts.Visibility, fh.Filename, err)
		return "", utils.PrependMessageToError(err, "failed to upload file")
	}
	return key, nil
}

func (s *objectStorage) SaveBytes(ctx context.Context, data []byte, opts SaveOptions) (string, error) {
	key := buildKey(opts, opts.Filename)
	contentType := opts.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}
	if opts.Visibility == VisibilityPublic {
		input.ACL = types.ObjectCannedACLPublicRead
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		log.Printf("cloud SaveBytes PutObject failed: bucket=%s key=%s contentType=%s visibility=%d bytes=%d err=%v", s.bucket, key, contentType, opts.Visibility, len(data), err)
		return "", utils.PrependMessageToError(err, "failed to upload bytes")
	}
	return key, nil
}

func (s *objectStorage) PublicURL(objectKey string) string {
	if objectKey == "" {
		return ""
	}
	return s.publicBaseURL + "/" + strings.TrimLeft(objectKey, "/")
}

func (s *objectStorage) PresignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *objectStorage) DeleteObject(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func buildKey(opts SaveOptions, srcFilename string) string {
	name := opts.Filename
	if name == "" {
		ext := strings.ToLower(filepath.Ext(srcFilename))
		name = uuid.New().String() + ext
	}
	ns := strings.Trim(opts.Namespace, "/")
	if ns == "" {
		return name
	}
	return ns + "/" + name
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
