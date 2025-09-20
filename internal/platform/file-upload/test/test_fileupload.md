# File Upload Module Test Plan

## Purpose & Scope
- Test Cloudflare R2 file upload functionality in `internal/platform/file-upload`
- Verify image processing integration, signed URL generation, and file management
- Cover error handling, edge cases, and performance requirements
- Ensure proper integration with image processing pipeline

## Component Map
- **R2Service (`r2.go`)**: Main implementation using AWS S3 SDK for Cloudflare R2
- **Uploader Interface (`port.go`)**: Contract defining upload, signed URL, and delete operations
- **Dependencies**: `imageprocess.Processor`, `environment.R2Environment`, AWS S3 client

## Requirements & Constraints
1. **File Processing**: All uploads go through image processing pipeline
2. **Key Generation**: Unique keys with UUID + timestamp format
3. **Signed URLs**: Configurable expiration times
4. **Error Handling**: Proper error propagation from dependencies
5. **Performance**: Upload operations should complete within reasonable time

## Test Strategy

### Unit Tests — R2Service
- Framework: `testing`, `testify/assert`, `testify/require`, `testify/mock`
- Dependencies: Mock AWS S3 client, mock image processor, test environment config

### Mock Dependencies Design

#### Mock S3 Client
- Implement interface matching AWS S3 client operations
- Track method calls for verification
- Configurable success/failure responses
- Support for testing different error scenarios

#### Mock Image Processor
- Return predictable `ProcessedImage` structs
- Configurable processing errors
- Track processing calls and parameters

#### Test Environment
- Use test-specific R2 configuration
- Avoid real Cloudflare R2 calls in unit tests

### Test Matrix

#### Constructor Tests
- **`NewR2Service`**
  - Valid environment creates service successfully
  - Nil processor defaults to WebpProcessor
  - AWS config creation with proper credentials
  - Base endpoint configuration for Cloudflare R2
  - Error handling for invalid AWS config

#### Upload Functionality
- **`UploadFile`**
  - Happy path: process image → generate key → upload to R2 → return key
  - Image processing failure propagates error
  - S3 upload failure returns error
  - Key generation format validation (UUID + timestamp)
  - Content type preservation from processed image
  - Context cancellation handling

#### Signed URL Generation
- **`GetSignedURL`**
  - Valid key returns signed URL
  - Expiration time configuration
  - Invalid key returns error
  - Context cancellation handling
  - URL format validation

#### File Deletion
- **`DeleteFile`**
  - Valid key deletes successfully
  - Invalid key returns error
  - Context cancellation handling

### Integration Tests — Real R2 Service
- Framework: `testing`, `testify/require`, `testcontainers-go` (if needed)
- Use test R2 bucket with cleanup procedures
- Real image processing pipeline

#### Integration Test Matrix
- **End-to-End Upload Flow**
  - Upload real image file
  - Verify file exists in R2 bucket
  - Generate signed URL and verify accessibility
  - Delete file and verify removal
- **Error Scenarios**
  - Invalid credentials
  - Network failures
  - Bucket access issues
- **Performance Tests**
  - Large file upload times
  - Concurrent upload handling
  - Memory usage during operations

### Test Utilities & Layout
```
internal/platform/file-upload/test/
├── r2_service_test.go        # Unit tests with mocks
├── integration_test.go       # Real R2 integration tests
├── mocks/
│   ├── s3_client_mock.go     # Mock S3 client implementation
│   └── processor_mock.go     # Mock image processor
├── testdata/
│   ├── test_image.jpg        # Sample image for testing
│   ├── test_image.png        # Different format test
│   └── large_image.jpg       # Performance testing
└── test_helpers.go           # Shared test utilities
```

### Mock Implementation Details

#### S3 Client Mock
```go
type MockS3Client struct {
    PutObjectFunc    func(ctx context.Context, input *s3.PutObjectInput) (*s3.PutObjectOutput, error)
    PresignGetObjectFunc func(ctx context.Context, input *s3.GetObjectInput) (*s3.PresignGetObjectOutput, error)
    DeleteObjectFunc func(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
    
    // Call tracking
    PutObjectCalls    []*s3.PutObjectInput
    PresignCalls      []*s3.GetObjectInput
    DeleteCalls       []*s3.DeleteObjectInput
}
```

#### Image Processor Mock
```go
type MockProcessor struct {
    ProcessFunc func(ctx context.Context, file io.Reader) (*imageprocess.ProcessedImage, error)
    ProcessCalls []io.Reader
}
```

### Test Scenarios

#### Happy Path Tests
- Upload JPEG image → WebP processing → R2 storage → key returned
- Generate signed URL with 1-hour expiration
- Delete uploaded file successfully

#### Error Path Tests
- Image processing failure (corrupted file)
- S3 upload failure (network error)
- Invalid credentials
- Context timeout/cancellation
- Invalid key for signed URL generation

#### Edge Cases
- Empty file upload
- Very large file handling
- Concurrent uploads
- Key collision handling (UUID uniqueness)
- Special characters in generated keys

#### Performance Tests
- Upload time benchmarks for different file sizes
- Memory usage during processing
- Concurrent upload limits
- Signed URL generation performance

### Environment Setup
- Test environment variables for R2 credentials
- Separate test bucket to avoid production data
- Cleanup procedures for test artifacts
- CI/CD integration with test credentials

### Running The Suite
- Unit tests: `go test ./internal/platform/file-upload/test -run Unit`
- Integration tests: `go test -tags=integration ./internal/platform/file-upload/test -run Integration`
- Performance tests: `go test -bench=. ./internal/platform/file-upload/test`
- Full suite: `make test-fileupload`

### Success Criteria
- ✅ All unit tests pass with mocked dependencies
- ✅ Integration tests work with real R2 service
- ✅ Error scenarios properly handled and tested
- ✅ Performance requirements met (upload < 5s for 10MB files)
- ✅ Memory usage stays reasonable (< 200MB per operation)
- ✅ Concurrent uploads handled correctly
- ✅ Proper cleanup of test artifacts

### Future Enhancements
- Add retry logic testing for transient failures
- Implement upload progress tracking tests
- Add metrics collection and monitoring tests
- Test with different image formats and edge cases
- Add security tests for signed URL validation
