// Package netx contains small HTTP helpers for interacting with presigned S3 URLs.
// It provides thin wrappers to upload and download binary blobs using standard
// net/http without bringing in an AWS SDK dependency.
package netx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// UploadToS3PresignedURL uploads raw bytes to a presigned S3 URL using HTTP PUT.
// The request sets Content-Type to "application/octet-stream". A non-200 status
// is treated as an error, and the response body (if any) is included for context.
func UploadToS3PresignedURL(url string, file []byte) error {
	req, err := http.NewRequest("PUT", url, bytes.NewReader(file))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s; body: %s", resp.Status, string(b))
	}
	return nil
}

// DownloadFromS3PresignedURL downloads bytes from a presigned S3 URL using HTTP GET.
// It accepts "application/octet-stream" and returns the full body for 200/206
// responses. Any other status is returned as an error with the response body text.
func DownloadFromS3PresignedURL(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusPartialContent:
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body failed: %w", err)
		}
		return data, nil
	default:
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed: %s; body: %s", resp.Status, string(b))
	}
}
