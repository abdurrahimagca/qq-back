package app

import (
	"context"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	"github.com/abdurrahimagca/qq-back/internal/ports"
)

type FileUsecase interface {
	GetSignedUrlForKey(ctx context.Context, key string, expires time.Duration) (*string, error)
}

type fileUsecase struct {
	environment       environment.Environment
	fileUploadService ports.UploadFilePort
}

func NewFileUsecase(fileUploadService ports.UploadFilePort, environment environment.Environment) FileUsecase {
	return &fileUsecase{fileUploadService: fileUploadService, environment: environment}
}

func (uc *fileUsecase) GetSignedUrlForKey(ctx context.Context, key string, expires time.Duration) (*string, error) {
	return uc.fileUploadService.GetSignedUrlByKey(ctx, key, expires)
}
