package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/dmitrijs2005/gophkeeper/internal/filex"
	"github.com/dmitrijs2005/gophkeeper/internal/netx"
)

// addEntry is a small workflow helper that:
//  1. prompts for the common "envelope" fields (title, metadata) and the
//     concrete entry payload via addEntryDetails,
//  2. (optionally) materializes a temporary file if the payload implements
//     models.Materializer,
//  3. delegates the final persist/sync to entryService.Add.
//
// On any failure the error is logged and returned unchanged.
func (a *App) addEntry(ctx context.Context, addEntryDetails func(context.Context) (models.TypedEntry, error)) error {
	item, file, err := a.InputEnvelope(ctx, a.reader, addEntryDetails)
	if err != nil {
		log.Printf("error: %v", err)
		return err
	}
	if err := a.entryService.Add(ctx, item, file, a.masterKey); err != nil {
		log.Printf("error: %v", err)
		return err
	}
	return nil
}

// AddNote collects a note body and persists it as a new entry.
func (a *App) AddNote(ctx context.Context) error {
	return a.addEntry(ctx, a.addNoteDetails)
}

// AddCreditCard collects credit-card fields and persists them as a new entry.
func (a *App) AddCreditCard(ctx context.Context) error {
	return a.addEntry(ctx, a.addCreditCardDetails)
}

// AddLogin collects login credentials and persists them as a new entry.
func (a *App) AddLogin(ctx context.Context) error {
	return a.addEntry(ctx, a.addLoginDetails)
}

// AddFile collects a file path and persists it as a new binary-file entry.
// The concrete file content is materialized by the payload when needed.
func (a *App) AddFile(ctx context.Context) error {
	return a.addEntry(ctx, a.addFileDetails)
}

// addNoteDetails prompts for a multi-line note text and returns a typed payload.
func (a *App) addNoteDetails(ctx context.Context) (models.TypedEntry, error) {
	text, err := GetMultiline(a.reader, "Enter note text (double Enter to finish):", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	return &models.Note{Text: text}, nil
}

// addCreditCardDetails prompts for card details and returns a typed payload.
func (a *App) addCreditCardDetails(ctx context.Context) (models.TypedEntry, error) {
	number, err := GetSimpleText(a.reader, "Enter card number", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	expiration, err := GetSimpleText(a.reader, "Enter expiration", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	return &models.CreditCard{Number: number, Expiration: expiration}, nil
}

// addLoginDetails prompts for login credentials and returns a typed payload.
func (a *App) addLoginDetails(ctx context.Context) (models.TypedEntry, error) {
	username, err := GetSimpleText(a.reader, "Enter username", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	password, err := GetSimpleText(a.reader, "Enter password", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	url, err := GetSimpleText(a.reader, "Enter URL", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	return &models.Login{Username: username, Password: password, URL: url}, nil
}

// addFileDetails prompts for a local file path and returns a typed payload.
func (a *App) addFileDetails(ctx context.Context) (models.TypedEntry, error) {
	filePath, err := GetSimpleText(a.reader, "Enter file path", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}
	return &models.BinaryFile{Path: filePath}, nil
}

// InputEnvelope gathers the common envelope data (title, metadata) and obtains
// a typed payload via 'rest'. If the payload implements models.Materializer,
// it is materialized into a *models.File (e.g., for binary uploads).
//
// Returns the constructed envelope, an optional *models.File, and an error.
func (a *App) InputEnvelope(
	ctx context.Context,
	r *bufio.Reader,
	rest func(ctx context.Context) (models.TypedEntry, error),
) (models.Envelope, *models.File, error) {
	var zero models.Envelope

	title, err := GetSimpleText(r, "Enter title", os.Stdout)
	if err != nil {
		return zero, nil, fmt.Errorf("get title: %w", err)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return zero, nil, fmt.Errorf("title is required")
	}
	if err := ctx.Err(); err != nil {
		return zero, nil, err
	}

	payload, err := rest(ctx)
	if err != nil {
		return zero, nil, err
	}

	var file *models.File
	if m, ok := payload.(models.Materializer); ok {
		file, err = m.Materialize(ctx)
		if err != nil {
			return zero, nil, fmt.Errorf("materialize: %w", err)
		}
	}

	md, err := GetMetadata(r)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, nil, err
	}
	metadata, err := models.MetadataFromString(md)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, nil, err
	}

	x, err := models.Wrap(payload.GetType(), title, metadata, payload)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, nil, err
	}
	return x, file, nil
}

// List prints a short textual representation for each stored entry.
// Decryption uses the in-memory master key.
func (a *App) List(ctx context.Context) error {
	s, err := a.entryService.List(ctx, a.masterKey)
	if err != nil {
		return err
	}
	for _, item := range s {
		fmt.Println(item)
	}
	return nil
}

// Sync triggers a two-way synchronization with the backend (if applicable).
func (a *App) Sync(ctx context.Context) error {
	return a.entryService.Sync(ctx)
}

// Delete removes an entry by its identifier, prompting the user for the ID.
func (a *App) Delete(ctx context.Context) error {
	id, err := GetSimpleText(a.reader, "Enter record id to delete", os.Stdout)
	if err != nil {
		return err
	}
	return a.entryService.DeleteByID(ctx, id)
}

// Show fetches and displays a single entry by ID.
//
// For structured types, it logs their fields. For binary files, it:
//  1. requests a presigned GET URL,
//  2. downloads the encrypted content,
//  3. fetches the per-file key/nonce,
//  4. ensures a local "download" directory,
//  5. decrypts the content to that directory using the original filename,
//  6. prints the destination path.
//
// Finally, it prints the envelope metadata as "name: value" lines.
func (a *App) Show(ctx context.Context) error {
	id, err := GetSimpleText(a.reader, "Enter record id to show", os.Stdout)
	if err != nil {
		return err
	}

	envelope, err := a.entryService.Get(ctx, id, a.masterKey)
	if err != nil {
		return err
	}

	log.Println(envelope.Title)

	x, err := envelope.Unwrap()
	if err != nil {
		return err
	}

	switch item := x.(type) {
	case models.Note:
		log.Printf("Note: %s", item.Text)

	case models.CreditCard:
		log.Printf("Number: %s", item.Number)
		log.Printf("Expiration: %s", item.Expiration)
		log.Printf("CVV: %s", item.CVV)
		log.Printf("Holder: %s", item.Holder)

	case models.Login:
		log.Printf("Username: %s", item.Username)
		log.Printf("Password: %s", item.Password)
		log.Printf("URL: %s", item.URL)

	case models.BinaryFile:
		// Download + decrypt the file to ./download/<basename>
		url, err := a.entryService.GetPresignedGetUrl(ctx, id)
		if err != nil {
			return err
		}

		encrypted, err := netx.DownloadFromS3PresignedURL(url)
		if err != nil {
			return err
		}

		fd, err := a.entryService.GetFile(ctx, id)
		if err != nil {
			return err
		}

		dir, err := filex.EnsureSubdDir("download")
		if err != nil {
			return err
		}

		ef := &cryptox.EncryptedFile{
			Cyphertext: encrypted,
			Key:        fd.EncryptedFileKey,
			Nonce:      fd.Nonce,
		}

		outputFile := filepath.Join(dir, filepath.Base(item.Path))

		if err := cryptox.DecryptFileTo(outputFile, ef); err != nil {
			return err
		}
		log.Printf("File saved to: %s", outputFile)
	}

	for _, md := range envelope.Metadata {
		log.Printf("%s: %s", md.Name, md.Value)
	}
	return nil
}
