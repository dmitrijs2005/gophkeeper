package netx

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

func TestDownloadFromS3PresignedURL_OK(t *testing.T) {
	t.Parallel()

	want := []byte("hello world")

	var seenAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	got, err := DownloadFromS3PresignedURL(srv.URL)
	require.NoError(t, err)
	require.Equal(t, want, got)
	require.Equal(t, "application/octet-stream", seenAccept, "Accept header must be set")
}

func TestDownloadFromS3PresignedURL_PartialContent(t *testing.T) {
	t.Parallel()

	want := []byte("partial-body")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// эмулируем 206 Partial Content
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	got, err := DownloadFromS3PresignedURL(srv.URL)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDownloadFromS3PresignedURL_ErrorStatus(t *testing.T) {
	t.Parallel()

	// типичный ответ S3/MinIO на ошибку
	xmlErr := `<?xml version="1.0" encoding="UTF-8"?><Error><Code>AccessDenied</Code></Error>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, xmlErr)
	}))
	defer srv.Close()

	_, err := DownloadFromS3PresignedURL(srv.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "403 Forbidden")
	require.Contains(t, err.Error(), "AccessDenied")
}

func TestDownloadFromS3PresignedURL_NetworkError(t *testing.T) {
	t.Parallel()

	// поднимаем сервер и сразу закрываем — получим connection error
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()

	_, err := DownloadFromS3PresignedURL(srv.URL)
	require.Error(t, err, "expected network error after server closed")
}

func TestDownloadFromS3PresignedURL_LargeBody(t *testing.T) {
	t.Parallel()

	// проверим, что читаем поток полностью
	large := make([]byte, 256*1024) // 256 KiB
	for i := range large {
		large[i] = byte(i % 251)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(large)
	}))
	defer srv.Close()

	got, err := DownloadFromS3PresignedURL(srv.URL)
	require.NoError(t, err)
	require.Equal(t, large, got)
}
