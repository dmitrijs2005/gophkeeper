// Package models defines vault entry types and their fields.
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

// EntryType classifies an entry kind.
type EntryType string

const (
	EntryTypeNote       EntryType = "note"
	EntryTypeBinaryFile EntryType = "binaryfile"
	EntryTypeLogin      EntryType = "login"
	EntryTypeCreditCard EntryType = "credit_card"
)

var ErrIncorrectMetadata = errors.New("metadata item must be name=value")

// ItemMetadata is a simple key/value pair.
type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

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

type Overview struct {
	Type  EntryType `json:"type"`
	Title string    `json:"title"`
}

type Envelope struct {
	Type     EntryType       `json:"type"`
	Title    string          `json:"title"`
	Metadata []Metadata      `json:"metadata"`
	Details  json.RawMessage `json:"details"`
}

func Wrap[T any](t EntryType, title string, md []Metadata, v T) (Envelope, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: t, Title: title, Metadata: md, Details: b}, nil
}

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

func (e Envelope) Overview() Overview {
	return Overview{Type: e.Type, Title: e.Title}
}

type TypedEntry interface {
	GetType() EntryType
}

// Login stores credentials.
type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
}

func (x Login) GetType() EntryType { return EntryTypeLogin }

// Note stores free-form text.
type Note struct {
	Text string `json:"text"`
}

func (x Note) GetType() EntryType { return EntryTypeNote }

// CreditCard stores payment card details.
type CreditCard struct {
	Number     string `json:"number"`
	Expiration string `json:"expiration"`
	CVV        string `json:"cvv"`
	Holder     string `json:"holder"`
}

func (x CreditCard) GetType() EntryType { return EntryTypeCreditCard }

// File references an encrypted file in external storage.
type BinaryFile struct {
	Path string `json:"path"`
}

func (f BinaryFile) GetType() EntryType { return EntryTypeBinaryFile }

type Materializer interface {
	Materialize(ctx context.Context) (*File, error)
}

func (f BinaryFile) Materialize(ctx context.Context) (*File, error) {

	dir, err := filex.EnsureSubdDir("preupload")
	if err != nil {
		return nil, fmt.Errorf("error creating dir: %w", err)
	}

	ef, err := cryptox.EncryptFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("error encrypting file: %w", err)
	}

	fn := fmt.Sprintf("%v", uuid.New())

	localPath := filepath.Join(dir, fn)

	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(ef.Cyphertext)
	if err != nil {
		return nil, fmt.Errorf("error writing temporary file: %w", err)
	}

	return &File{EncryptedFileKey: ef.Key, Nonce: ef.Nonce, LocalPath: localPath}, nil
}
