// Command gophkeeper-server boots the GophKeeper backend service.
//
// It loads runtime configuration, initializes a structured JSON logger,
// constructs the application from a database DSN, and starts the server.
// On initialization failure the process logs the error and exits.
package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	"github.com/dmitrijs2005/gophkeeper/internal/server"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
)

// main wires configuration, logging, and the server application, then runs it.
//
// Startup sequence:
//  1. Load configuration (see internal/server/config).
//  2. Create a JSON slog logger and wrap it with the project's logging facade.
//  3. Build the server application using the configured DSN.
//  4. Run the application until it stops or fails.
//
// Any construction error is logged; the process then exits with a non-zero code.
func main() {
	cfg := config.LoadConfig()

	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := logging.NewSlogLogger(l)

	app, err := server.NewAppFromDSN(cfg, logger)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	app.Run()
}
