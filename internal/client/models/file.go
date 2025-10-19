package models

// FileUploadTask describes a pending server request for the client to upload
// a file associated with an entry.
type FileUploadTask struct {
	// EntryID is the ID of the entry whose file must be uploaded.
	EntryID string
	// URL is a presigned (temporary) PUT/POST URL to upload the encrypted blob.
	URL string
}

// File stores metadata for a file associated with an entry.
//
// EncryptedFileKey and Nonce are the per-file encryption materials (client-side
// encrypted content is stored remotely). LocalPath points to a local ciphertext
// file when present (e.g., during pre-upload). UploadStatus can be used by the
// sync layer to track progress; values are application-defined.
type File struct {
	// EntryID links this file to its parent entry.
	EntryID string
	// EncryptedFileKey is the (encrypted) symmetric key for the file contents.
	EncryptedFileKey []byte
	// Nonce is the AEAD nonce used for file content encryption.
	Nonce []byte
	// LocalPath is a path to a locally stored ciphertext (temporary/staging).
	LocalPath string
	// UploadStatus indicates client-side upload progress, e.g. "pending"/"completed".
	UploadStatus string
	// Deleted marks the file as a tombstone for synchronization/GC.
	Deleted bool
}
