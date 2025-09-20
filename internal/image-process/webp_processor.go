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
	"math"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
)

const (
	defaultMaxWidth  = 2048
	defaultMaxHeight = 1080
	defaultMaxPixels = 400000
	defaultQuality   = 45
	defaultMethod    = 0
)

type WebpProcessor struct {
	maxWidth  uint
	maxHeight uint
	maxPixels uint
	quality   int
	method    int
}

var ErrWebpProcessorUnavailable = errors.New("webp processing requires cgo support")

func NewWebpProcessor() *WebpProcessor {
	return &WebpProcessor{
		maxWidth:  defaultMaxWidth,
		maxHeight: defaultMaxHeight,
		maxPixels: defaultMaxPixels,
		quality:   defaultQuality,
		method:    defaultMethod,
	}
}

func (p *WebpProcessor) ImageProcessor(ctx context.Context, file io.Reader) (*ProcessedImage, error) {
	_ = ctx
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	processed := img
	imgBounds := img.Bounds()
	width := uint(imgBounds.Dx())
	height := uint(imgBounds.Dy())

	if width > 0 && height > 0 && (width > p.maxWidth || height > p.maxHeight) {
		scale := math.Min(float64(p.maxWidth)/float64(width), float64(p.maxHeight)/float64(height))
		if scale > 1 {
			scale = 1
		}
		newWidth := uint(math.Max(1, math.Round(float64(width)*scale)))
		newHeight := uint(math.Max(1, math.Round(float64(height)*scale)))

		if p.maxPixels > 0 {
			targetPixels := newWidth * newHeight
			if targetPixels > p.maxPixels {
				pixelScale := math.Sqrt(float64(p.maxPixels) / float64(targetPixels))
				newWidth = uint(math.Max(1, math.Round(float64(newWidth)*pixelScale)))
				newHeight = uint(math.Max(1, math.Round(float64(newHeight)*pixelScale)))
			}
		}
		processed = resize.Resize(newWidth, newHeight, img, resize.NearestNeighbor)
	}

	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(p.quality))
	if err != nil {
		return nil, err
	}

	options.Method = p.method
	options.ThreadLevel = true

	if err = webp.Encode(&buf, processed, options); err != nil {
		return nil, err
	}

	return &ProcessedImage{
		Data:     buf.Bytes(),
		MimeType: "image/webp",
	}, nil
}
