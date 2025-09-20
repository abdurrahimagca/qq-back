package imageprocess

import (
	"context"
	"io"
)

type ProcessedImage struct {
	Data     []byte
	MimeType string
}

type Processor interface {
	ImageProcessor(ctx context.Context, file io.Reader) (*ProcessedImage, error)
}
