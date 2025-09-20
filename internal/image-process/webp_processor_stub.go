//go:build !cgo

package imageprocess

import (
	"context"
	"errors"
	"io"
)

var ErrWebpProcessorUnavailable = errors.New("webp processing requires cgo support")

type WebpProcessor struct{}

func NewWebpProcessor() *WebpProcessor {
	return &WebpProcessor{}
}

func (p *WebpProcessor) ImageProcessor(ctx context.Context, file io.Reader) (*ProcessedImage, error) {
	_ = ctx
	_ = file
	return nil, ErrWebpProcessorUnavailable
}
