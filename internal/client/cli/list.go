package cli

import (
	"context"
	"fmt"
	"log"
)

func (a *App) list(ctx context.Context) {
	s, err := a.entryService.List(ctx, a.masterKey)
	if err != nil {
		log.Printf(err.Error())
	}

	for _, item := range s {
		fmt.Println(item)
	}
}
