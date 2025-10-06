package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	var c Config
	c.LoadDefaults()

	assert.Equal(t, c.DatabaseDSN, "postgres://postgres:postgres@postgres:5432/gophkeeper?sslmode=disable")
	assert.Equal(t, c.EndpointAddrGRPC, ":50051")
	assert.Equal(t, c.SecretKey, "secretKey")
	assert.Equal(t, c.AccessTokenValidityDuration, 1*time.Minute)
	assert.Equal(t, c.RefreshTokenValidityDuration, 3*time.Minute)
	assert.Equal(t, c.S3RootUser, "admin")
	assert.Equal(t, c.S3RootPassword, "secretpassword")
	assert.Equal(t, c.S3Bucket, "vault")
	assert.Equal(t, c.S3Region, "us-east-1")
	assert.Equal(t, c.S3BaseEndpoint, "http://127.0.0.1:9000/")
}

func TestLoadConfig_UsesDefaultsBeforeParsing(t *testing.T) {
	c := LoadConfig()

	require.NotNil(t, c, "LoadConfig must not return nil")

	assert.Equal(t, c.DatabaseDSN, "postgres://postgres:postgres@postgres:5432/gophkeeper?sslmode=disable")
	assert.Equal(t, c.EndpointAddrGRPC, ":50051")
	assert.Equal(t, c.SecretKey, "secretKey")
	assert.Equal(t, c.AccessTokenValidityDuration, 1*time.Minute)
	assert.Equal(t, c.RefreshTokenValidityDuration, 3*time.Minute)
	assert.Equal(t, c.S3RootUser, "admin")
	assert.Equal(t, c.S3RootPassword, "secretpassword")
	assert.Equal(t, c.S3Bucket, "vault")
	assert.Equal(t, c.S3Region, "us-east-1")
	assert.Equal(t, c.S3BaseEndpoint, "http://127.0.0.1:9000/")
}
