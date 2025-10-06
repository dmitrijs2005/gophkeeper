package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/flagx"
)

// c.EndpointAddrGRPC = ":50051"
// c.DatabaseDSN = "postgres://postgres:postgres@postgres:5432/gophkeeper?sslmode=disable"
// c.SecretKey = "secretKey"
// c.AccessTokenValidityDuration = 1 * time.Minute
// c.RefreshTokenValidityDuration = 3 * time.Minute
// c.S3RootUser = "admin"
// c.S3RootPassword = "secretpassword"
// c.S3Bucket = "vault"
// c.S3Region = "us-east-1"
// c.S3BaseEndpoint = "http://127.0.0.1:9000/"

func parseFlags(config *Config) {

	// filtering args to leave just values processed by parseFlags
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

	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	config.AccessTokenValidityDuration = time.Duration(*accessTokenValidityDuration) * time.Minute
	config.RefreshTokenValidityDuration = time.Duration(*refreshTokenValidityDuration) * time.Minute

}
