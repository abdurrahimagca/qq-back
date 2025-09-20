package app

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	fileupload "github.com/abdurrahimagca/qq-back/internal/platform/file-upload"
)

type FileUsecase interface {
	GetSignedUrlForKey(ctx context.Context, key string, expires time.Duration) (*string, error)
}

type fileUsecase struct {
	environment       environment.Environment
	fileUploadService fileupload.Uploader
}

func NewFileUsecase(fileUploadService fileupload.Uploader, environment environment.Environment) FileUsecase {
	return &fileUsecase{fileUploadService: fileUploadService, environment: environment}
}

func (uc *fileUsecase) GetSignedUrlForKey(ctx context.Context, key string, expires time.Duration) (*string, error) {
	return uc.fileUploadService.GetSignedURL(ctx, key, expires)
}
