package fileupload

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	imageprocess "github.com/abdurrahimagca/qq-back/internal/image-process"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2Service struct {
	client      *s3.Client
	environment environment.R2Environment
	processor   imageprocess.Processor
}

func NewR2Service(environment environment.R2Environment, processor imageprocess.Processor) Uploader {
	accessKeyID := environment.AccessKeyID
	accessKeySecret := environment.SecretAccessKey
	accountID := environment.AccountID
	logger := slog.Default()

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		logger.Error("Error creating AWS config", "error", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
	})

	if processor == nil {
		processor = imageprocess.NewWebpProcessor()
	}

	return &R2Service{
		client:      client,
		environment: environment,
		processor:   processor,
	}
}

func (s *R2Service) UploadFile(ctx context.Context, file io.Reader) (*string, error) {
	processedImage, err := s.processor.ImageProcessor(ctx, file)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s-%s", uuid.New().String(), time.Now().Format("2006-01-02"))
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.environment.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(processedImage.Data),
		ContentType: aws.String(processedImage.MimeType),
	})

	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *R2Service) GetSignedURL(ctx context.Context, key string, expires time.Duration) (*string, error) {
	presignClient := s3.NewPresignClient(s.client)
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		ResponseExpires: aws.Time(time.Now().Add(expires)),
		Bucket:          aws.String(s.environment.BucketName),
		Key:             aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &presignResult.URL, nil
}

func (s *R2Service) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.environment.BucketName),
		Key:    aws.String(key),
	})
	return err
}
