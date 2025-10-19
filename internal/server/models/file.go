// Package models defines server-side data models persisted in the database.
package models

// File describes server-side metadata for a binary payload associated
// with an entry. The encrypted content itself is stored in object storage.
type File struct {
	// EntryID links the file to its parent entry.
	EntryID string
	// UserID is the owner of the file.
	UserID string
	// Version is the server-assigned, monotonic version used for sync.
	Version int64

	// StorageKey is the object-storage key (path) of the ciphertext blob.
	StorageKey string
	// EncryptedFileKey is the per-file symmetric key (itself encrypted, if applicable).
	EncryptedFileKey []byte
	// Nonce is the AEAD nonce used to encrypt the file contents.
	Nonce []byte

	// UploadStatus tracks server-side upload state (e.g., "pending", "completed").
	UploadStatus string
}

// FileUploadTask instructs the client to upload a file using a presigned URL.
type FileUploadTask struct {
	// EntryID identifies which entry's file should be uploaded.
	EntryID string
	// URL is a temporary presigned HTTP URL for the client to PUT the ciphertext.
	URL string
}
