package bucket

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type UploadImageResult struct {
	Key *string
	isSuccess bool
}

type Service struct {
	client *s3.Client
	config environment.R2Config
}

func NewService(cfg environment.R2Config) (*Service, error) {
	accessKeyId := cfg.AccessKeyID
	accessKeySecret := cfg.SecretAccessKey
	accountId := cfg.AccountID

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	

	return &Service{
		client: client,
		config: cfg,
	}, nil
}

func (s *Service) UploadImage(file io.Reader, uploadImagePublic bool) (UploadImageResult, error) {
	totalStart := time.Now()
	
	processedImage, err := ProcessSingleImage(file)
	if err != nil {
		return UploadImageResult{}, err
	}

	uploadStart := time.Now()
	log.Printf("Starting single image upload...")

	key := fmt.Sprintf("%s-%s", uuid.New().String(), time.Now().Format("2006-01-02"))

	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(processedImage.Data),
		ContentType: aws.String(processedImage.MimeType),
	})

	if err != nil {
		return UploadImageResult{}, err
	}


	log.Printf("Total upload time: %v", time.Since(uploadStart))
	
	uploadResult := UploadImageResult{
		Key:    &key, // Use Medium field for single image
		isSuccess: true,
	}
	
	log.Printf("Total operation time: %v", time.Since(totalStart))
	return uploadResult, nil
}

func (s *Service) GetPresignedURL(key string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	presignResult, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		ResponseExpires: aws.Time(time.Now().Add(expiresIn)),
		Bucket:          aws.String(s.config.BucketName),
		Key:             aws.String(key),
	})
	if err != nil {
		return "", err
	}
	return presignResult.URL, nil
}

func (s *Service) DeleteImage(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(key),
	})
	return err
}
