//go:build cgo

package imageprocess

import (
	"bytes"
	"context"
	"errors"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
)

const (
	defaultMaxWidth  = 2048
	defaultMaxHeight = 1080
	defaultQuality   = 85
)

type WebpProcessor struct {
	maxWidth  uint
	maxHeight uint
	quality   int
}

var ErrWebpProcessorUnavailable = errors.New("webp processing requires cgo support")

func NewWebpProcessor() *WebpProcessor {
	return &WebpProcessor{
		maxWidth:  defaultMaxWidth,
		maxHeight: defaultMaxHeight,
		quality:   defaultQuality,
	}
}

func (p *WebpProcessor) ImageProcessor(ctx context.Context, file io.Reader) (*ProcessedImage, error) {
	_ = ctx
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	thumb := resize.Thumbnail(p.maxWidth, p.maxHeight, img, resize.Lanczos3)

	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(p.quality))
	if err != nil {
		return nil, err
	}

	if err = webp.Encode(&buf, thumb, options); err != nil {
		return nil, err
	}

	return &ProcessedImage{
		Data:     buf.Bytes(),
		MimeType: "image/webp",
	}, nil
}
