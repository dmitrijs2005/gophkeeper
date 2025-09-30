package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	"github.com/dmitrijs2005/gophkeeper/internal/server"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
)

func main() {

	ctx := context.Background()
	cfg := config.LoadConfig()

	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := logging.NewSlogLogger(l)

	app, err := server.NewAppFromDSN(cfg, logger)

	if err != nil {
		log.Printf("%v", err)
		return
	}

	app.Run(ctx)

}
