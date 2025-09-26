package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/models"
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

	fmt.Println(envelope.Title)

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
	}

	for _, md := range envelope.Metadata {
		log.Printf("%s: %s", md.Name, md.Value)
	}

}
