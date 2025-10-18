package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
)

// parseFlags populates selected Config fields from command-line flags.
//
// Supported flags (short forms):
//
//	-a string   address and port of the backend server (default from Config)
//	-i int      online check interval in seconds (default from Config)
//
// Note: The function filters os.Args to only include the flags it knows about,
// using flagx.FilterArgs, to avoid interference with other components.
func parseFlags(cfg *Config) {
	// Filter args to include only those handled here.
	args := flagx.FilterArgs(os.Args[1:], []string{"-a", "-i"})

	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&cfg.ServerEndpointAddr, "a", cfg.ServerEndpointAddr, "address and port to access server")
	onlineCheckInterval := fs.Int("i", int(cfg.OnlineCheckInterval.Seconds()), "online check interval (in seconds)")

	if err := fs.Parse(args); err != nil {
		panic(err)
	}

	cfg.OnlineCheckInterval = time.Duration(*onlineCheckInterval) * time.Second
}
