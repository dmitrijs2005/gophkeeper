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

func (a *App) addEntry(ctx context.Context, addEntryDetails func(context.Context) (models.TypedEntry, error)) {

	item, file, err := a.InputEnvelope(ctx, a.reader, addEntryDetails)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	err = a.entryService.Add(ctx, item, file, a.masterKey)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

}

func (a *App) addNote(ctx context.Context) {
	a.addEntry(ctx, a.addNoteDetails)
}

func (a *App) addCreditCard(ctx context.Context) {
	a.addEntry(ctx, a.addCreditCardDetails)
}

func (a *App) addLogin(ctx context.Context) {
	a.addEntry(ctx, a.addLoginDetails)
}

func (a *App) addFile(ctx context.Context) {
	a.addEntry(ctx, a.addFileDetails)
}

func (a *App) addNoteDetails(ctx context.Context) (models.TypedEntry, error) {

	text, err := GetMultiline(a.reader, "Enter note text (double Enter to finish):", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.Note{Text: text}, nil

}

func (a *App) addCreditCardDetails(ctx context.Context) (models.TypedEntry, error) {

	number, err := GetSimpleText(a.reader, "Enter card number", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	expiration, err := GetSimpleText(a.reader, "Enter card number", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.CreditCard{Number: number, Expiration: expiration}, nil

}

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

func (a *App) addFileDetails(ctx context.Context) (models.TypedEntry, error) {

	filePath, err := GetSimpleText(a.reader, "Enter file path", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.BinaryFile{Path: filePath}, nil

}

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

func (a *App) list(ctx context.Context) {
	s, err := a.entryService.List(ctx, a.masterKey)
	if err != nil {
		log.Println(err.Error())
	}

	for _, item := range s {
		fmt.Println(item)
	}
}

func (a *App) sync(ctx context.Context) {
	err := a.entryService.Sync(ctx)
	if err != nil {
		log.Println(err.Error())
	}

}

func (a *App) delete(ctx context.Context) {

	id, err := GetSimpleText(a.reader, "Enter record id to delete", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	err = a.entryService.DeleteByID(ctx, id)

	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

func (a *App) show(ctx context.Context) {

	id, err := GetSimpleText(a.reader, "Enter record id to show", os.Stdout)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	envelope, err := a.entryService.Get(ctx, id, a.masterKey)

	if err != nil {
		log.Printf("Error: %s", err.Error())
	}

	log.Println(envelope.Title)

	x, err := envelope.Unwrap()
	if err != nil {
		log.Printf("Error: %s", err.Error())
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

		url, err := a.entryService.GetPresignedGetUrl(ctx, id)

		if err != nil {
			log.Printf("Error getting download url: %s", err.Error())
			return
		}

		encrypted, err := netx.DownloadFromS3PresignedURL(url)
		if err != nil {
			log.Printf("Error downloading file: %s", err.Error())
			return
		}

		fd, err := a.entryService.GetFile(ctx, id)
		if err != nil {
			log.Printf("Error getting file details: %s", err.Error())
			return
		}

		dir, err := filex.EnsureSubdDir("download")
		if err != nil {
			log.Printf("Error creating dir: %s", err.Error())
			return
		}

		ef := &cryptox.EncryptedFile{Cyphertext: encrypted, Key: fd.EncryptedFileKey, Nonce: fd.Nonce}

		ouputFile := filepath.Join(dir, filepath.Base(item.Path))

		err = cryptox.DecryptFileTo(ouputFile, ef)
		if err != nil {
			log.Printf("Error creating dir: %s", err.Error())
			return
		}

		log.Printf("File saved to: %s", ouputFile)
	}

	for _, md := range envelope.Metadata {
		log.Printf("%s: %s", md.Name, md.Value)
	}

}
