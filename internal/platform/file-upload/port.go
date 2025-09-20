package fileupload

import (
	"context"
	"io"
	"time"
)

// Uploader provides the contract for persisting files and retrieving signed URLs.
type Uploader interface {
	UploadFile(ctx context.Context, file io.Reader) (*string, error)
	GetSignedURL(ctx context.Context, key string, expires time.Duration) (*string, error)
}
