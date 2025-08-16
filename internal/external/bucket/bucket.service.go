package bucket

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"sync"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/config/environment"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type UploadImageResult struct {
	Small     *UploadImageResultItem
	Medium    *UploadImageResultItem
	Large     *UploadImageResultItem
	isSuccess bool
}

type UploadImageResultItem struct {
	Key       string
	PublicURL *string
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

func (s *Service) UploadImage(file multipart.File, imageType ImageTypeTarget, uploadImagePublic bool) (UploadImageResult, error) {
	imageServiceResult, err := ProcessImage(file, imageType, []ImageVariant{SMALL, MEDIUM, LARGE})
	if err != nil {
		return UploadImageResult{}, err
	}

	uploadResult := UploadImageResult{}

	processAndUpload := func(image *ProcessedImage) (*UploadImageResultItem, error) {
		if image == nil {
			return nil, nil
		}

		key := fmt.Sprintf("%s-%s", uuid.New().String(), time.Now().Format("2006-01-02"))

		_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      aws.String(s.config.BucketName),
			Key:         aws.String(key),
			Body:        bytes.NewReader(image.Data),
			ContentType: aws.String(image.MimeType),
		})

		if err != nil {
			return nil, err
		}

		resultItem := &UploadImageResultItem{
			Key: key,
		}

		if uploadImagePublic {
			publicURL := fmt.Sprintf("%s/%s", s.config.URL, key)
			resultItem.PublicURL = &publicURL
		}

		return resultItem, nil
	}

	// Upload all variants in parallel
	var wg sync.WaitGroup
	var uploadErr error
	var mu sync.Mutex

	if imageServiceResult.Small != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := processAndUpload(imageServiceResult.Small)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				uploadErr = err
				return
			}
			uploadResult.Small = result
		}()
	}

	if imageServiceResult.Medium != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := processAndUpload(imageServiceResult.Medium)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				uploadErr = err
				return
			}
			uploadResult.Medium = result
		}()
	}

	if imageServiceResult.Large != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := processAndUpload(imageServiceResult.Large)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				uploadErr = err
				return
			}
			uploadResult.Large = result
		}()
	}

	wg.Wait()

	if uploadErr != nil {
		return UploadImageResult{}, uploadErr
	}

	uploadResult.isSuccess = true
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
