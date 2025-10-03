package main

import (
	"context"
	"log"

	"github.com/dmitrijs2005/gophkeeper/internal/client/cli"
	"github.com/dmitrijs2005/gophkeeper/internal/client/config"
)

func main() {

	ctx := context.Background()
	cfg := config.LoadConfig()
	app, err := cli.NewApp(cfg)

	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	app.Run(ctx)

}
