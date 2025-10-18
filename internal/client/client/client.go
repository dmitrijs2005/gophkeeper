package client

import (
	"context"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

// Client is the high-level contract for server interaction:
// account bootstrap, authentication, liveness checks, two-way sync,
// and file transfer helpers (e.g., presigned URLs).
//
// Implementations should be safe for concurrent use unless stated otherwise.
type Client interface {
	// Close releases any underlying resources (connections, goroutines, etc.).
	Close() error

	// Register creates a new account using a username, salt, and derived key.
	// The salt/key are typically produced client-side (e.g., from a password KDF).
	Register(ctx context.Context, username string, salt []byte, key []byte) error

	// GetSalt returns the server-stored salt for the given username,
	// used to derive the authentication key locally.
	GetSalt(ctx context.Context, username string) ([]byte, error)

	// Login authenticates with the server using the derived key.
	// Subsequent calls may refresh tokens/credentials as needed.
	Login(ctx context.Context, username string, key []byte) error

	// Ping performs a lightweight reachability/liveness probe.
	Ping(ctx context.Context) error

	// Sync performs bidirectional synchronization.
	//   - entries/files: local changes to push
	//   - maxVersion:    caller's latest known version
	// It returns:
	//   - newOrUpdatedEntriesFromServer
	//   - deletedEntriesFromServer (if represented as tombstones)
	//   - newOrUpdatedFilesFromServer
	//   - pendingUploads (server requests client to upload these)
	//   - newMaxVersion (server's latest version after merge)
	Sync(ctx context.Context,
		entries []*models.Entry,
		files []*models.File,
		maxVersion int64,
	) (
		updatedEntries []*models.Entry,
		deletedEntries []*models.Entry,
		updatedFiles []*models.File,
		pendingUploads []*models.FileUploadTask,
		newMaxVersion int64,
		err error,
	)

	// MarkUploaded acknowledges to the server that a file for entryID
	// has been successfully uploaded.
	MarkUploaded(ctx context.Context, entryID string) error

	// GetPresignedGetURL returns a temporary, signed URL for downloading
	// an encrypted file associated with entryID.
	GetPresignedGetURL(ctx context.Context, entryID string) (string, error)
}
