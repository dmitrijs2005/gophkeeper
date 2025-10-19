package auth

// internal/auth/auth_test.go

import (
	"testing"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
)

func TestGenerateAndParse_Success(t *testing.T) {
	t.Parallel()

	secret := []byte("super-secret")
	userID := "user-123"

	tok, err := GenerateToken(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	gotUserID, err := GetUserIDFromToken(tok, secret)
	if err != nil {
		t.Fatalf("GetUserIDFromToken error: %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("userID mismatch: got %q want %q", gotUserID, userID)
	}
}

func TestGetUserIDFromToken_Expired(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")
	userID := "u1"

	tok, err := GenerateToken(userID, secret, -1*time.Second)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	_, err = GetUserIDFromToken(tok, secret)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
	if err != common.ErrTokenExpired {
		t.Fatalf("expected common.ErrTokenExpired, got %v", err)
	}
}

func TestGetUserIDFromToken_WrongSecret(t *testing.T) {
	t.Parallel()

	userID := "u2"
	tok, err := GenerateToken(userID, []byte("right-secret"), time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	_, err = GetUserIDFromToken(tok, []byte("wrong-secret"))
	if err == nil {
		t.Fatalf("expected error for invalid signature, got nil")
	}
}

func TestGetUserIDFromToken_MalformedString(t *testing.T) {
	t.Parallel()

	_, err := GetUserIDFromToken("not.a.jwt", []byte("k"))
	if err == nil {
		t.Fatalf("expected error for malformed token, got nil")
	}
}
