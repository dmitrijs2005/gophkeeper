// Package models defines vault entry types and their fields.
package models

import (
	"encoding/json"
	"errors"
	"strings"
)

// EntryType classifies an entry kind.
type EntryType string

const (
	EntryTypeNote        EntryType = "note"
	EntryTypeFile        EntryType = "file"
	EntryTypeLogin       EntryType = "login"
	EntryTypeCreditCard  EntryType = "credit_card"
	EntryTypeFileOffline EntryType = "file_offline"
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
	case EntryTypeFile:
		var v File
		return v, json.Unmarshal(e.Details, &v)
	default:
		return nil, json.Unmarshal(e.Details, &map[string]any{})
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
type File struct {
	StorageKey string `json:"storage_key"`
	FileKey    []byte `json:"file_key"`
	Nonce      []byte `json:"nonce"`
}

func (x File) GetType() EntryType { return EntryTypeFile }

// structure for offline file adding
type FileOffline struct {
	Path string `json:"path"`
}

func (x FileOffline) GetType() EntryType { return EntryTypeFileOffline }
