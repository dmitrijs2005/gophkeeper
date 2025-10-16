package cli

import (
	"context"
	"log"
	"path/filepath"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
	"github.com/dmitrijs2005/gophkeeper/internal/cryptox"
	"github.com/dmitrijs2005/gophkeeper/internal/filex"
	"github.com/dmitrijs2005/gophkeeper/internal/netx"
)

func (a *App) delete(ctx context.Context) {

	id, err := GetSimpleText(a.reader, "Enter record id to delete")
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

	id, err := GetSimpleText(a.reader, "Enter record id to show")
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
