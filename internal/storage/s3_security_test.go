package storage

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestDownloadSizeLimit(t *testing.T) {
	for _, chunked := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if chunked {
				w.(http.Flusher).Flush()
			} else {
				w.Header().Set("Content-Length", "4096")
			}
			_, _ = w.Write([]byte(strings.Repeat("x", 4096)))
		}))
		client := s3.NewFromConfig(aws.Config{Region: "us-east-1", Credentials: credentials.NewStaticCredentialsProvider("fixture", "fixture", "")}, func(o *s3.Options) { o.BaseEndpoint = aws.String(server.URL); o.UsePathStyle = true })
		store := &S3Storage{client: client, bucket: "test"}
		if _, err := store.Download(context.Background(), "backup.json", 1024); !errors.Is(err, ErrObjectTooLarge) {
			t.Errorf("chunked=%v: %v", chunked, err)
		}
		content, err := store.Download(context.Background(), "backup.json", 4096)
		if err != nil || len(content) != 4096 {
			t.Errorf("bounded valid download: %d, %v", len(content), err)
		}
		server.Close()
	}
}
