package cli

import (
	"context"
	"fmt"
	"log"
)

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
