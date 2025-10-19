package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
)

// parseFlags populates selected server Config fields from command-line flags.
//
// Supported flags (short forms):
//
//	-a string   gRPC bind address (e.g., ":50051")
//	-d string   PostgreSQL DSN
//	-s string   JWT HMAC secret key
//	-t int      access token validity, minutes
//	-r int      refresh token validity, minutes
//	-u string   S3 root user
//	-p string   S3 root password
//	-b string   S3 bucket name
//	-g string   S3 region
//	-e string   S3 base endpoint (e.g., "http://127.0.0.1:9000/")
//
// Notes:
//   - The function first filters os.Args to only the flags it recognizes using
//     flagx.FilterArgs, avoiding collisions with other components.
//   - Duration flags are accepted as integers in minutes and then converted
//     to time.Duration values.
func parseFlags(config *Config) {
	// Filter args to include only the flags handled here.
	args := flagx.FilterArgs(os.Args[1:], []string{"-a", "-d", "-s", "-t", "-r", "-u", "-p", "-b", "-g", "-e"})

	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&config.EndpointAddrGRPC, "a", config.EndpointAddrGRPC, "address and port to run server")
	fs.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database DSN")
	fs.StringVar(&config.SecretKey, "s", config.SecretKey, "secret key")

	accessTokenValidityDuration := fs.Int("t", int(config.AccessTokenValidityDuration.Minutes()), "access_token_validity_duration (in minutes)")
	refreshTokenValidityDuration := fs.Int("r", int(config.RefreshTokenValidityDuration.Minutes()), "refresh_token_validity_duration (in minutes)")

	fs.StringVar(&config.S3RootUser, "u", config.S3RootUser, "S3 root user")
	fs.StringVar(&config.S3RootPassword, "p", config.S3RootPassword, "S3 root password")
	fs.StringVar(&config.S3Bucket, "b", config.S3Bucket, "S3 root bucket")
	fs.StringVar(&config.S3Region, "g", config.S3Region, "S3 root region")
	fs.StringVar(&config.S3BaseEndpoint, "e", config.S3BaseEndpoint, "S3 base endpoint")

	if err := fs.Parse(args); err != nil {
		panic(err)
	}

	config.AccessTokenValidityDuration = time.Duration(*accessTokenValidityDuration) * time.Minute
	config.RefreshTokenValidityDuration = time.Duration(*refreshTokenValidityDuration) * time.Minute
}
