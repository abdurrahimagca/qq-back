package bucket

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"

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
	POST_SQUARE:    {SMALL: {Width: 1024, Height: 1024}, MEDIUM: {Width: 10424, Height: 1024}, LARGE: {Width: 1024, Height: 1024}},
	POST_LANDSCAPE: {SMALL: {Width: 1024, Height: 600}, MEDIUM: {Width: 1024, Height: 600}, LARGE: {Width: 1024, Height: 600}},
	POST_PORTRAIT:  {SMALL: {Width: 600, Height: 1024}, MEDIUM: {Width: 600, Height: 1024}, LARGE: {Width: 600, Height: 1024}},
	AVATAR:         {SMALL: {Width: 100, Height: 100}, MEDIUM: {Width: 100, Height: 100}, LARGE: {Width: 100, Height: 100}},
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
		// Create a new RGBA image with the target size
		dst := image.NewRGBA(image.Rect(0, 0, size.ImageSize.Width, size.ImageSize.Height))

		// Resize the image using CatmullRom interpolation
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

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
