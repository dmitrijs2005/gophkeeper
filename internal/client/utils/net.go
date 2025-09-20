package utils

import (
	"bytes"
	"fmt"
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
		return fmt.Errorf("upload failed: %s", resp.Status)
	}
	return nil
}
