package cloud

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	config "github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type awsService struct {
	service    *s3.Client
	bucketName string
}

const (
	filePreSignExpireDuration = time.Hour * 12
)

func NewAWSCloudService(cfg config.Config) (CloudService, error) {
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AwsRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsAccessKeyID, cfg.AwsSecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	svc := s3.NewFromConfig(awsCfg)

	return &awsService{
		service:    svc,
		bucketName: cfg.AwsBucketName,
	}, nil
}

// session, err := session.NewSession(&aws.Config{
//  Region:      aws.String(cfg.AwsRegion),
//  Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsSecretKey, ""),
// })
// if err != nil {
//  return nil, fmt.Errorf("failed to create session for aws service : %w", err)
// }

// service := s3.New(session)

// return &awsService{
//  service:    service,
//  bucketName: cfg.AwsBucketName,
// }, nil
// }

func (c *awsService) SaveFile(ctx context.Context, fileHeader *multipart.FileHeader) (string, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to open file")
	}

	uploadID := uuid.New().String()

	_, err = c.service.PutObject(ctx, &s3.PutObjectInput{
		Body:   file,
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(uploadID),
	})
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to upload file")
	}

	return uploadID, nil
}
func (c *awsService) GetFileUrl(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(c.service)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.bucketName,
		Key:    &key,
	}, s3.WithPresignExpires(12*time.Hour))

	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}

func UploadProfileImage(c *gin.Context, fileHeader *multipart.FileHeader) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}
	defer openedFile.Close() //https://pkg.go.dev/golang.org/x/tools/internal/typesinternal#UndeclaredImportedNamee()

	ctx := context.Background()
	// Load your config to get the bucket name
	// Replace this with your actual config initialization or pass config as a parameter
	appConfig := config.Config{
		AwsBucketName: "s3-mandi-bucket",
		AwsRegion:     "ap-south-1",
		// Add other required fields here
	}
	bucketName := appConfig.AwsBucketName

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(appConfig.AwsRegion))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load AWS config"})
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(s3Client)

	// Generate unique file name
	fileName := fmt.Sprintf("user-profile/%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))

	// Upload to S3
	contentType := file.Header.Get("Content-Type")
	result, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &fileName,
		Body:        openedFile,
		ContentType: &contentType, // pass address of variable
		ACL:         "public-read",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to S3"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Upload successful",
		"image_url": result.Location,
	})
}
