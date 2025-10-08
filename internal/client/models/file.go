// Package models defines vault entry types and their fields.
package models

type File struct {
	ID               string
	EntryID          string
	EncryptedFileKey []byte
	Nonce            []byte
	LocalPath        string
	UploadStatus     string
	Deleted          bool
}
