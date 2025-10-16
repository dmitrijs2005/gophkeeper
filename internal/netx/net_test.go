package netx

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadToS3PresignedURL(t *testing.T) {
	file := []byte("hello, s3")

	t.Run("success 200 OK", func(t *testing.T) {
		var gotBody []byte
		var gotCT string
		var gotMethod string

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotCT = r.Header.Get("Content-Type")
			body, _ := io.ReadAll(r.Body)
			_ = r.Body.Close()
			gotBody = body
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		err := UploadToS3PresignedURL(ts.URL+"/some/presigned?X-Amz-Signature=abc", file)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotMethod != http.MethodPut {
			t.Fatalf("method = %q, want PUT", gotMethod)
		}
		if gotCT != "application/octet-stream" {
			t.Fatalf("Content-Type = %q, want application/octet-stream", gotCT)
		}
		if !bytes.Equal(gotBody, file) {
			t.Fatalf("body = %q, want %q", string(gotBody), string(file))
		}
	})

	t.Run("non-200 -> error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden) // 403
		}))
		defer ts.Close()

		err := UploadToS3PresignedURL(ts.URL, file)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "upload failed: 403") {
			t.Fatalf("error = %q, want to contain 403", err.Error())
		}
	})

	t.Run("network error", func(t *testing.T) {
		ts := httptest.NewServer(http.NotFoundHandler())
		ts.Close()

		err := UploadToS3PresignedURL(ts.URL, file)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !isNetOpError(err) {
			if strings.Contains(err.Error(), "upload failed") {
				t.Fatalf("got wrong kind of error: %v", err)
			}
		}
	})
}

type netOpErrorLike interface {
	error
	Timeout() bool
	Temporary() bool
}

func isNetOpError(err error) bool {
	var target netOpErrorLike
	return errors.As(err, &target)
}
