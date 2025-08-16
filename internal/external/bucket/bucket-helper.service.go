package bucket

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"

	_ "github.com/adrium/goheif"
	"golang.org/x/image/draw"
)

type ProcessedImage struct {
	Data     []byte
	MimeType string
}

type ImageServiceResult struct {
	Small     *ProcessedImage
	Medium    *ProcessedImage
	Large     *ProcessedImage
	isSuccess bool
}

type ImageSize struct {
	Width  int
	Height int
}

type ImageConstraints struct {
	MaxWidth     *int
	MaxHeight    *int
	MinWidth     *int
	MinHeight    *int
	FallbackSize *ImageSize
}

type ImageTypeTarget string

const (
	POST_SQUARE    ImageTypeTarget = "POST_SQUARE"
	POST_LANDSCAPE ImageTypeTarget = "POST_LANDSCAPE"
	POST_PORTRAIT  ImageTypeTarget = "POST_PORTRAIT"
	AVATAR         ImageTypeTarget = "AVATAR"
)

type ImageVariant int

const (
	SMALL  ImageVariant = 1
	MEDIUM ImageVariant = 2
	LARGE  ImageVariant = 3
)

var TARGET_IMAGE_SIZES = map[ImageTypeTarget]map[ImageVariant]ImageSize{
	POST_SQUARE:    {SMALL: {Width: 600, Height: 600}, MEDIUM: {Width: 900, Height: 900}, LARGE: {Width: 1024, Height: 1024}},
	POST_LANDSCAPE: {SMALL: {Width: 1024, Height: 600}, MEDIUM: {Width: 1024, Height: 600}, LARGE: {Width: 1024, Height: 600}},
	POST_PORTRAIT:  {SMALL: {Width: 600, Height: 1024}, MEDIUM: {Width: 600, Height: 1024}, LARGE: {Width: 600, Height: 1024}},
	AVATAR:         {SMALL: {Width: 300, Height: 300}, MEDIUM: {Width: 600, Height: 600}, LARGE: {Width: 900, Height: 900}},
}

type ImageSizeWithVariant struct {
	ImageSize ImageSize
	Variant   ImageVariant
}

func getInternalImageSizesBasedOnImageTypeTargetWithVariants(imageType ImageTypeTarget, variants []ImageVariant) ([]ImageSizeWithVariant, error) {
	imageSizes := []ImageSizeWithVariant{}
	for _, variant := range variants {
		imageSizes = append(imageSizes, ImageSizeWithVariant{ImageSize: TARGET_IMAGE_SIZES[imageType][variant], Variant: variant})
	}

	return imageSizes, nil
}

func qualityBasedOnVariant(variant ImageVariant) int {
	switch variant {
	case SMALL:
		return 60
	case MEDIUM:
		return 70
	case LARGE:
		return 90
	}
	return 90
}

func ProcessImage(file multipart.File, imageType ImageTypeTarget, variants []ImageVariant) (ImageServiceResult, error) {
	imageSizes, err := getInternalImageSizesBasedOnImageTypeTargetWithVariants(imageType, variants)
	if err != nil {
		return ImageServiceResult{}, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return ImageServiceResult{}, err
	}

	result := ImageServiceResult{}

	for _, size := range imageSizes {
		// Calculate crop dimensions to maintain aspect ratio
		srcBounds := img.Bounds()
		srcWidth := srcBounds.Dx()
		srcHeight := srcBounds.Dy()

		targetWidth := size.ImageSize.Width
		targetHeight := size.ImageSize.Height

		// Calculate scale factors for both dimensions
		scaleX := float64(targetWidth) / float64(srcWidth)
		scaleY := float64(targetHeight) / float64(srcHeight)

		// Use the larger scale factor to ensure the image covers the entire target area
		scale := scaleX
		if scaleY > scaleX {
			scale = scaleY
		}

		// Calculate the scaled source dimensions
		scaledWidth := int(float64(srcWidth) * scale)
		scaledHeight := int(float64(srcHeight) * scale)

		// Calculate crop offsets to center the image
		cropX := (scaledWidth - targetWidth) / 2
		cropY := (scaledHeight - targetHeight) / 2

		// Create intermediate scaled image
		scaledImg := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))
		draw.CatmullRom.Scale(scaledImg, scaledImg.Bounds(), img, img.Bounds(), draw.Over, nil)

		// Create final cropped image
		dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

		// Copy the cropped portion
		cropRect := image.Rect(cropX, cropY, cropX+targetWidth, cropY+targetHeight)
		draw.Draw(dst, dst.Bounds(), scaledImg, cropRect.Min, draw.Src)

		// Encode the resized image to JPEG format
		var buf bytes.Buffer
		err := jpeg.Encode(&buf, dst, &jpeg.Options{
			Quality: qualityBasedOnVariant(size.Variant),
		})
		if err != nil {
			return ImageServiceResult{}, err
		}

		processedImage := &ProcessedImage{
			Data:     buf.Bytes(),
			MimeType: "image/jpeg",
		}

		switch size.Variant {
		case SMALL:
			result.Small = processedImage
		case MEDIUM:
			result.Medium = processedImage
		case LARGE:
			result.Large = processedImage
		}
	}

	result.isSuccess = true
	return result, nil
}
