package main

import (
	"context"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/server"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
)

func main() {

	ctx := context.Background()
	cfg := config.LoadConfig()
	app, err := server.NewApp(cfg)

	if err != nil {
		log.Printf("%v", err)
		return
	}

	app.Run(ctx)

}
