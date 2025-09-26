package cli

import (
	"context"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
)

func (a *App) addNote(ctx context.Context) {

	title, err := GetSimpleText(a.reader, "- Enter title")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	text, err := GetMultiline(a.reader, "- Enter note text (double Enter to finish):")
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	metadata, err := GetMetadata(a.reader)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	details := models.Note{Text: text}
	m, err := models.MetadataFromString(metadata)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}

	x, err := models.Wrap(models.EntryTypeNote, title, m, details)
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
