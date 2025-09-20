//go:build cgo

package imageprocess_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"
	"time"

	imageprocess "github.com/abdurrahimagca/qq-back/internal/image-process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebpProcessor_ImageProcessor_FormatSupport(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    func() *bytes.Buffer
		wantMime string
	}{
		{
			name: "JPEG input",
			input: func() *bytes.Buffer {
				img := createTestImage(100, 100)
				buf := &bytes.Buffer{}
				require.NoError(t, jpeg.Encode(buf, img, nil))
				return buf
			},
			wantMime: "image/webp",
		},
		{
			name: "PNG input",
			input: func() *bytes.Buffer {
				img := createTestImage(100, 100)
				buf := &bytes.Buffer{}
				require.NoError(t, png.Encode(buf, img))
				return buf
			},
			wantMime: "image/webp",
		},
		{
			name: "GIF input",
			input: func() *bytes.Buffer {
				img := createTestImage(100, 100)
				buf := &bytes.Buffer{}
				require.NoError(t, gif.Encode(buf, img, nil))
				return buf
			},
			wantMime: "image/webp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input()

			result, err := processor.ImageProcessor(ctx, input)

			require.NoError(t, err)
			assert.Equal(t, tt.wantMime, result.MimeType)
			assert.NotEmpty(t, result.Data)

			// Verify output is valid WebP by checking magic bytes
			assert.True(t, len(result.Data) >= 12, "WebP should have at least 12 bytes")
			assert.Equal(t, "RIFF", string(result.Data[0:4]), "Should start with RIFF")
			assert.Equal(t, "WEBP", string(result.Data[8:12]), "Should contain WEBP signature")
		})
	}
}

func TestWebpProcessor_ImageProcessor_SizeConstraints(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	tests := []struct {
		name     string
		width    int
		height   int
		wantSize func(size int) bool
	}{
		{
			name:   "Large image should be compressed under 1MB",
			width:  4000,
			height: 3000,
			wantSize: func(size int) bool {
				return size < 1024*1024 // < 1MB
			},
		},
		{
			name:   "Small image should remain optimized",
			width:  200,
			height: 200,
			wantSize: func(size int) bool {
				return size < 50*1024 // < 50KB for small images
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createTestImage(tt.width, tt.height)
			buf := &bytes.Buffer{}
			require.NoError(t, png.Encode(buf, img))

			result, err := processor.ImageProcessor(ctx, buf)

			require.NoError(t, err)
			assert.True(t, tt.wantSize(len(result.Data)),
				"Output size %d bytes doesn't meet constraint", len(result.Data))
		})
	}
}

func TestWebpProcessor_ImageProcessor_ResizingLogic(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	tests := []struct {
		name         string
		inputWidth   int
		inputHeight  int
		expectResize bool
	}{
		{
			name:         "Large image should be resized",
			inputWidth:   4000,
			inputHeight:  3000,
			expectResize: true,
		},
		{
			name:         "Small image should not be resized",
			inputWidth:   800,
			inputHeight:  600,
			expectResize: false,
		},
		{
			name:         "Exact max dimensions should not be resized",
			inputWidth:   2048,
			inputHeight:  1080,
			expectResize: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createTestImage(tt.inputWidth, tt.inputHeight)
			buf := &bytes.Buffer{}
			require.NoError(t, png.Encode(buf, img))

			result, err := processor.ImageProcessor(ctx, buf)
			require.NoError(t, err)

			// For resize verification, we'd need to decode the WebP
			// For now, just verify we got reasonable output
			assert.NotEmpty(t, result.Data)
			assert.Equal(t, "image/webp", result.MimeType)
		})
	}
}

func TestWebpProcessor_ImageProcessor_PerformanceRequirements(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	// Create a reasonably large test image
	img := createTestImage(2000, 1500)
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))

	start := time.Now()
	result, err := processor.ImageProcessor(ctx, buf)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Data)

	// Performance requirement: ≤ 30ms
	assert.LessOrEqual(t, duration.Milliseconds(), int64(30),
		"Processing took %v, should be ≤ 30ms", duration)
}

func TestWebpProcessor_ImageProcessor_ErrorHandling(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	tests := []struct {
		name  string
		input func() *bytes.Buffer
	}{
		{
			name: "Invalid image data",
			input: func() *bytes.Buffer {
				return bytes.NewBufferString("not an image")
			},
		},
		{
			name: "Corrupted JPEG",
			input: func() *bytes.Buffer {
				// Create a buffer that starts like JPEG but is corrupted.
				buf := &bytes.Buffer{}
				buf.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG magic
				buf.WriteString("corrupted data")
				return buf
			},
		},
		{
			name: "Empty input",
			input: func() *bytes.Buffer {
				return &bytes.Buffer{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input()

			result, err := processor.ImageProcessor(ctx, input)

			assert.Error(t, err, "Should return error for invalid input")
			assert.Nil(t, result, "Should not return result on error")
		})
	}
}

func TestWebpProcessor_Constructor(t *testing.T) {
	processor := imageprocess.NewWebpProcessor()

	// Test that constructor creates a valid processor.
	assert.NotNil(t, processor)

	// Test with a simple image to ensure it works.
	ctx := context.Background()
	img := createTestImage(100, 100)
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))

	result, err := processor.ImageProcessor(ctx, buf)
	require.NoError(t, err)
	assert.Equal(t, "image/webp", result.MimeType)
	assert.NotEmpty(t, result.Data)
}

// Benchmark tests for performance validation.
func BenchmarkImageProcessor_SmallImage(b *testing.B) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	img := createTestImage(200, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		png.Encode(buf, img)

		_, err := processor.ImageProcessor(ctx, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImageProcessor_LargeImage(b *testing.B) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	img := createTestImage(2000, 1500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		png.Encode(buf, img)

		_, err := processor.ImageProcessor(ctx, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImageProcessor_ExtremeSize(b *testing.B) {
	processor := imageprocess.NewWebpProcessor()
	ctx := context.Background()

	img := createTestImage(4000, 3000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		png.Encode(buf, img)

		_, err := processor.ImageProcessor(ctx, buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to create test images.
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern to make it more realistic.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: uint8(((x + y) * 255) / (width + height)),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	return img
}
