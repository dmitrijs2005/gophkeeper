package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

func InputEnvelope[T models.TypedEntry](
	ctx context.Context,
	r *bufio.Reader,
	rest func(ctx context.Context) (T, error),
) (models.Envelope, error) {

	var zero models.Envelope
	title, err := GetSimpleText(r, "Enter title")
	if err != nil {
		return zero, fmt.Errorf("get title: %w", err)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return zero, fmt.Errorf("title is required")
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	payload, err := rest(ctx)
	if err != nil {
		return zero, err
	}

	md, err := GetMetadata(r)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, err
	}

	metadata, err := models.MetadataFromString(md)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, err
	}

	x, err := models.Wrap(payload.GetType(), title, metadata, payload)
	if err != nil {
		log.Printf("error: %v", err)
		return zero, err
	}

	return x, nil
}

func (a *App) addCreditCard(ctx context.Context) {

	item, err := InputEnvelope(ctx, a.reader, a.addCreditCardDetails)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	err = a.entryService.Add(ctx, item, a.masterKey)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

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

func (a *App) addLogin(ctx context.Context) {

	title, err := GetSimpleText(a.reader, "Enter title")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	username, err := GetSimpleText(a.reader, "Enter username")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	password, err := GetSimpleText(a.reader, "Enter password")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	url, err := GetSimpleText(a.reader, "Enter URL")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	md, err := GetMetadata(a.reader)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	details := models.Login{Username: username, Password: password, URL: url}
	metadata, err := models.MetadataFromString(md)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	x, err := models.Wrap(models.EntryTypeLogin, title, metadata, details)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	err = a.entryService.Add(ctx, x, a.masterKey)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

}
