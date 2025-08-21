package ports

import (
	"context"
	"io"
	"time"
)

type UploadFilePort interface {
	UploadFile(ctx context.Context, file io.Reader) (*string, error)
	GetSignedUrlByKey(ctx context.Context, key string, expires time.Duration) (*string, error)
}
