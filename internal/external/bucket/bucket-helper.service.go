package bucket

import (
	"bytes"
	"image"
	_ "image/gif"
		_ "image/jpeg"
	_ "image/png"
	"log"
	"mime/multipart"
	"time"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
)

type ProcessedImage struct {
	Data     []byte
	MimeType string
}

func ProcessSingleImage(file multipart.File) (*ProcessedImage, error) {
	start := time.Now()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Create thumbnail
	thumb := resize.Thumbnail(2048, 1080, img, resize.Lanczos3)

	// Encode to webp
	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 80)
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
