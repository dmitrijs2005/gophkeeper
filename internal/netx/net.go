package netx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

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
