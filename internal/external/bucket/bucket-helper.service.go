package bucket

import (
	_ "image/gif"
	_ "image/png"
	"log"
	"mime/multipart"
	"time"

	"github.com/cshum/vipsgen/vips"
)

type ProcessedImage struct {
	Data     []byte
	MimeType string
}

func ProcessSingleImage(file multipart.File) (*ProcessedImage, error) {
	start := time.Now()

	// Decode image
	source := vips.NewSource(file)
	defer source.Close()

	image, err := vips.NewThumbnailSource(source, 800, &vips.ThumbnailSourceOptions{
		FailOn: vips.FailOnError, // Fail on first error
	})

	webpData, err := image.WebpsaveBuffer(&vips.WebpsaveBufferOptions{
		Q: 85, // Quality factor (0-100)

		SmartSubsample: true, // Better chroma subsampling
	})
	if err != nil {
		return nil, err
	}
	processedImage := &ProcessedImage{
		Data:     webpData,
		MimeType: "image/webp",
	}

	log.Printf("Total single image processing took: %v", time.Since(start))
	return processedImage, nil

}
