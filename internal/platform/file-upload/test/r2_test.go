package fileupload_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/environment"
	imageprocess "github.com/abdurrahimagca/qq-back/internal/image-process"
	fileupload "github.com/abdurrahimagca/qq-back/internal/platform/file-upload"
)

type failingProcessor struct{}

func (f *failingProcessor) ImageProcessor(ctx context.Context, r io.Reader) (*imageprocess.ProcessedImage, error) {
	return nil, errors.New("processor failed")
}

func TestGetSignedURL_ReturnsURL(t *testing.T) {
	env := environment.R2Environment{
		BucketName:      "test-bucket",
		URL:             "https://example.com",
		TokenValue:      "token",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		AccountID:       "acc",
	}

	svc := fileupload.NewR2Service(env, nil)

	url, err := svc.GetSignedURL(context.Background(), "test-key", time.Minute)
	if err != nil {
		t.Fatalf("GetSignedURL returned error: %v", err)
	}
	if url == nil || *url == "" {
		t.Fatalf("expected non-empty URL, got %v", url)
	}
	if !strings.Contains(*url, "test-key") {
		t.Fatalf("signed URL does not contain key: %s", *url)
	}
}

func TestUploadFile_ProcessorError(t *testing.T) {
	env := environment.R2Environment{
		BucketName:      "test-bucket",
		URL:             "https://example.com",
		TokenValue:      "token",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
		AccountID:       "acc",
	}

	svc := fileupload.NewR2Service(env, &failingProcessor{})

	key, err := svc.UploadFile(context.Background(), strings.NewReader("input"))
	if err == nil {
		t.Fatalf("expected error from UploadFile when processor fails, got nil")
	}
	if key != nil {
		t.Fatalf("expected nil key when processor fails, got %v", *key)
	}
}
