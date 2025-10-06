package main

import (
	"context"
	"log"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/buildinfo"
	"github.com/dmitrijs2005/gophkeeper/internal/client/cli"
	"github.com/dmitrijs2005/gophkeeper/internal/client/config"
)

func main() {

	buildinfo.PrintBuildData(os.Stdout)

	ctx := context.Background()
	cfg := config.LoadConfig()
	app, err := cli.NewApp(cfg)

	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	app.Run(ctx)

}
