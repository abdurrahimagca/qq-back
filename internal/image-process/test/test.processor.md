
# Image Process Module Test Plan

## Purpose & Scope
- Test image compression and processing functionality in `internal/image-process`
- Verify performance requirements and output quality
- Cover WebP conversion, resizing, and compression logic

## Component Map
- **Service (`service.go`)**: Defines `Processor` interface and `ProcessedImage` struct
- **WebpProcessor (`webp_processor.go`)**: Core implementation that converts images to WebP format with compression and resizing

## Requirements & Constraints
1. **Compression**: Output images must be < 1MB
2. **Performance**: Processing time must be ≤ 30ms even for large files
3. **Format Support**: Handle JPEG, PNG, GIF inputs → WebP output
4. **Quality**: Maintain acceptable visual quality while meeting size constraints

## Test Strategy

### Unit Tests — WebpProcessor
- Framework: `testing`, `testify/assert`, `testify/require`
- Test image samples: embed test images in `testdata/` directory

### Test Matrix

#### Core Functionality
- **`ImageProcessor` - Format Support**
  - JPEG input → WebP output with correct MIME type
  - PNG input → WebP output (transparency handling)
  - GIF input → WebP output (static frame)
  - Invalid/corrupted image → proper error handling

#### Compression Requirements
- **Size Constraint Validation**
  - Large images (>5MB) → output < 1MB
  - Small images remain optimized
  - Quality setting produces expected compression ratios

#### Performance Requirements
- **Processing Time**
  - Measure processing duration for various image sizes
  - Large files (10MB+) complete within 30ms
  - Use `testing.B` for benchmark tests
  - Profile memory usage to avoid excessive allocations

#### Image Quality
- **Resizing Logic**
  - Images larger than 2048x1080 get resized proportionally
  - Aspect ratio preservation
  - Smaller images remain unchanged
  - Edge cases: very wide/tall aspect ratios

#### Configuration
- **WebpProcessor Settings**
  - Default quality (85) produces expected results
  - Max dimensions (2048x1080) enforced correctly
  - Constructor sets proper defaults

### Test Data Setup
```
internal/image-process/test/
├── webp_processor_test.go     # Main test implementation
├── test.processor.md          # This documentation
└── testdata/
    ├── large_photo.jpg        # >5MB test image
    ├── small_icon.png         # <100KB test image
    ├── wide_banner.png        # Extreme aspect ratio
    ├── transparent.png        # PNG with transparency
    ├── animated.gif           # GIF for static conversion
    └── corrupted.jpg          # Invalid image data
```

### Performance Benchmarks
- `BenchmarkImageProcessor_SmallImage`
- `BenchmarkImageProcessor_LargeImage`
- `BenchmarkImageProcessor_ExtremeSize`

### Error Handling Tests
- Invalid image format
- Corrupted image data
- WebP encoding failures
- Context cancellation (if implemented)

## Running Tests
- Unit tests: `go test ./internal/image-process/test`
- Benchmarks: `go test -bench=. ./internal/image-process/test`
- With race detection: `go test -race ./internal/image-process/test`

## Success Criteria
- ✅ All input formats convert successfully to WebP
- ✅ Output size < 1MB for all test images
- ✅ Processing time < 30ms per image
- ✅ Visual quality acceptable (manual verification)
- ✅ Memory usage stays reasonable (< 100MB per operation)