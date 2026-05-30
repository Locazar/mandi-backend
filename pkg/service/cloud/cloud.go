package cloud

import (
	"context"
	"mime/multipart"
	"time"
)

type Visibility int

const (
	VisibilityPublic Visibility = iota
	VisibilityPrivate
)

type SaveOptions struct {
	Namespace   string
	Visibility  Visibility
	ContentType string
	Filename    string
}

type CloudService interface {
	SaveFile(ctx context.Context, fileHeader *multipart.FileHeader, opts SaveOptions) (objectKey string, err error)
	SaveBytes(ctx context.Context, data []byte, opts SaveOptions) (objectKey string, err error)
	PublicURL(objectKey string) string
	PresignedURL(ctx context.Context, objectKey string, ttl time.Duration) (url string, err error)
	DeleteObject(ctx context.Context, objectKey string) error
}
