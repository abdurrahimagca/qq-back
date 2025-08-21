package fileupload

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/ports"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
)

type ProcessedImage struct {
	Data     []byte
	MimeType string
}

type R2Service struct {
	client      *s3.Client
	environment environment.R2Environment
}

func NewR2Service(environment environment.R2Environment) ports.UploadFilePort {
	accessKeyId := environment.AccessKeyID
	accessKeySecret := environment.SecretAccessKey
	accountId := environment.AccountID

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

	return &R2Service{
		client:      client,
		environment: environment,
	}
}

func (s *R2Service) UploadFile(ctx context.Context, file io.Reader) (*string, error) {
	totalStart := time.Now()

	processedImage, err := s.processSingleImage(file)
	if err != nil {
		return nil, err
	}

	uploadStart := time.Now()
	log.Printf("Starting single image upload...")

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

	log.Printf("Total upload time: %v", time.Since(uploadStart))
	log.Printf("Total operation time: %v", time.Since(totalStart))

	return &key, nil
}

func (s *R2Service) GetSignedUrlByKey(ctx context.Context, key string, expires time.Duration) (*string, error) {
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

func (s *R2Service) processSingleImage(file io.Reader) (*ProcessedImage, error) {
	start := time.Now()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	thumb := resize.Thumbnail(2048, 1080, img, resize.Lanczos3)

	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 85)
	if err != nil {
		return nil, err
	}

	if err := webp.Encode(&buf, thumb, options); err != nil {
		return nil, err
	}

	processedImage := &ProcessedImage{
		Data:     buf.Bytes(),
		MimeType: "image/webp",
	}

	log.Printf("Total single image processing took: %v", time.Since(start))
	return processedImage, nil
}
