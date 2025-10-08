package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
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

	text, err := GetMultiline(a.reader, "- Enter note text (double Enter to finish):")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.Note{Text: text}, nil

}

func (a *App) addCreditCardDetails(ctx context.Context) (models.TypedEntry, error) {

	number, err := GetSimpleText(a.reader, "Enter card number")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	expiration, err := GetSimpleText(a.reader, "Enter card number")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.CreditCard{Number: number, Expiration: expiration}, nil

}

func (a *App) addLoginDetails(ctx context.Context) (models.TypedEntry, error) {

	username, err := GetSimpleText(a.reader, "Enter username")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	password, err := GetSimpleText(a.reader, "Enter password")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	url, err := GetSimpleText(a.reader, "Enter URL")
	if err != nil {
		log.Printf("error: %v", err)
		return nil, err
	}

	return &models.Login{Username: username, Password: password, URL: url}, nil

}

func (a *App) addFileDetails(ctx context.Context) (models.TypedEntry, error) {

	filePath, err := GetSimpleText(a.reader, "Enter file path")
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

	title, err := GetSimpleText(r, "Enter title")
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
