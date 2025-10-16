package models

type File struct {
	EntryID          string
	UserID           string
	Version          int64
	StorageKey       string
	EncryptedFileKey []byte
	Nonce            []byte
	UploadStatus     string
}

type FileUploadTask struct {
	EntryID string
	URL     string
}
