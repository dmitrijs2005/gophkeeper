package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/dmitrijs2005/gophkeeper/internal/filex"
	"github.com/google/uuid"
)

// EntryType classifies the kind of entry stored in the vault.
type EntryType string

const (
	EntryTypeNote       EntryType = "note"
	EntryTypeBinaryFile EntryType = "binaryfile"
	EntryTypeLogin      EntryType = "login"
	EntryTypeCreditCard EntryType = "credit_card"
)

// ErrIncorrectMetadata is returned when a metadata line is not "name=value".
var ErrIncorrectMetadata = errors.New("metadata item must be name=value")

// Metadata is a simple key/value pair attached to an entry.
type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MetadataFromString converts lines of "name=value" into []Metadata.
func MetadataFromString(s []string) ([]Metadata, error) {
	data := make([]Metadata, len(s))
	for n, item := range s {
		parts := strings.Split(item, "=")
		if len(parts) != 2 {
			return nil, ErrIncorrectMetadata
		}
		data[n] = Metadata{Name: parts[0], Value: parts[1]}
	}
	return data, nil
}

// Overview is a compact summary for listing/searching.
type Overview struct {
	Type  EntryType `json:"type"`
	Title string    `json:"title"`
}

// Envelope is a typed, JSON-serializable wrapper around a concrete entry
// payload placed into Details as raw JSON bytes.
type Envelope struct {
	Type     EntryType       `json:"type"`
	Title    string          `json:"title"`
	Metadata []Metadata      `json:"metadata"`
	Details  json.RawMessage `json:"details"`
}

// Wrap marshals v into Details and returns an Envelope with type and metadata.
func Wrap[T any](t EntryType, title string, md []Metadata, v T) (Envelope, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: t, Title: title, Metadata: md, Details: b}, nil
}

// Unwrap decodes Details into a concrete typed struct based on Envelope.Type.
// Unknown types are returned as a generic map[string]any.
func (e Envelope) Unwrap() (any, error) {
	switch e.Type {
	case EntryTypeLogin:
		var v Login
		return v, json.Unmarshal(e.Details, &v)
	case EntryTypeNote:
		var v Note
		return v, json.Unmarshal(e.Details, &v)
	case EntryTypeCreditCard:
		var v CreditCard
		return v, json.Unmarshal(e.Details, &v)
	case EntryTypeBinaryFile:
		var v BinaryFile
		return v, json.Unmarshal(e.Details, &v)
	default:
		var m map[string]any
		if err := json.Unmarshal(e.Details, &m); err != nil {
			return nil, err
		}
		return m, nil
	}
}

// Overview returns a compact summary object for the envelope.
func (e Envelope) Overview() Overview {
	return Overview{Type: e.Type, Title: e.Title}
}

// TypedEntry is implemented by all concrete entry payloads to report their type.
type TypedEntry interface {
	GetType() EntryType
}

// Login stores website/app credentials.
type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

func (x Login) GetType() EntryType { return EntryTypeLogin }

// Note stores free-form text content.
type Note struct {
	Text string `json:"text"`
}

func (x Note) GetType() EntryType { return EntryTypeNote }

// CreditCard stores payment card details (no formatting/validation enforced here).
type CreditCard struct {
	Number     string `json:"number"`
	Expiration string `json:"expiration"`
	CVV        string `json:"cvv"`
	Holder     string `json:"holder"`
}

func (x CreditCard) GetType() EntryType { return EntryTypeCreditCard }

// BinaryFile references a local file path to be encrypted and uploaded.
type BinaryFile struct {
	Path string `json:"path"`
}

func (f BinaryFile) GetType() EntryType { return EntryTypeBinaryFile }

// Materializer turns a typed value into a temporary encrypted file blob
// ready for upload (plus the per-file key/nonce).
type Materializer interface {
	Materialize(ctx context.Context) (*File, error)
}

// Materialize encrypts the file at BinaryFile.Path, writes its ciphertext to
// a unique file under the local "preupload" directory, and returns a *File
// containing the encrypted bytes' key/nonce and the temporary path.
//
// Returned File.LocalPath points to the ciphertext, not the plaintext.
func (f BinaryFile) Materialize(ctx context.Context) (*File, error) {
	dir, err := filex.EnsureSubdDir("preupload")
	if err != nil {
		return nil, fmt.Errorf("error creating dir: %w", err)
	}

	ef, err := cryptox.EncryptFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("error encrypting file: %w", err)
	}

	fn := uuid.New().String()
	localPath := filepath.Join(dir, fn)

	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	if _, err = file.Write(ef.Cyphertext); err != nil {
		return nil, fmt.Errorf("error writing temporary file: %w", err)
	}

	return &File{
		EncryptedFileKey: ef.Key,
		Nonce:            ef.Nonce,
		LocalPath:        localPath,
	}, nil
}
