// Command gophkeeper is a CLI client for the GophKeeper project.
// It prints build metadata, loads runtime configuration, constructs the CLI
// application, and starts its main event loop. See the internal packages
// for details on configuration shape and available commands.
package main

import (
	"context"
	"log"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/buildinfo"
	"github.com/dmitrijs2005/gophkeeper/internal/client/cli"
	"github.com/dmitrijs2005/gophkeeper/internal/client/config"
)

// main is the entry point of the CLI client.
//
// Execution flow:
//  1. Print build information (version, commit, build time) to stdout.
//  2. Load configuration from the environment and/or config files.
//  3. Construct the CLI App with the loaded configuration.
//  4. Run the application until it exits or an unrecoverable error occurs.
//
// On initialization failures the process terminates with a non-zero exit code.
// For configuration semantics refer to the config package; for interactive
// behavior and commands refer to the cli package.
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
