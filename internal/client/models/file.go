// Package models defines vault entry types and their fields.
package models

type FileUploadTask struct {
	EntryID string
	URL     string
}

type File struct {
	EntryID          string
	EncryptedFileKey []byte
	Nonce            []byte
	LocalPath        string
	UploadStatus     string
	Deleted          bool
}
