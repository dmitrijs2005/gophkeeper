package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
)

func parseFlags(config *Config) {

	// filtering args to leave just values processed by parseFlags
	args := flagx.FilterArgs(os.Args[1:], []string{"-a", "-i"})

	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&config.ServerEndpointAddr, "a", config.ServerEndpointAddr, "address and port to access server")
	onlineCheckInterval := fs.Int("i", int(config.OnlineCheckInterval.Seconds()), "online check interval (in seconds)")

	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	config.OnlineCheckInterval = time.Duration(*onlineCheckInterval) * time.Second

}
