package cryptox

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// ---------- helpers ----------

type sample struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func must[T any](t *testing.T, v T, err error) T {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return v
}

// ---------- MakeVerifier / DeriveMasterKey ----------

func TestDeriveMasterKey_DeterministicAndLength(t *testing.T) {
	pass := []byte("secret")
	salt := []byte("salty-salt")
	k1 := DeriveMasterKey(pass, salt)
	k2 := DeriveMasterKey(pass, salt)

	if !bytes.Equal(k1, k2) {
		t.Fatalf("expected deterministic output for same inputs")
	}
	if len(k1) != 32 {
		t.Fatalf("expected key length 32, got %d", len(k1))
	}
}

func TestMakeVerifier_ChangesWhenKeyChanges(t *testing.T) {
	k1 := bytes.Repeat([]byte{1}, 32)
	k2 := bytes.Repeat([]byte{2}, 32)
	v1 := MakeVerifier(k1)
	v2 := MakeVerifier(k2)
	if bytes.Equal(v1, v2) {
		t.Fatalf("verifiers for different keys should differ")
	}
	if len(v1) != 32 {
		t.Fatalf("expected verifier length 32, got %d", len(v1))
	}
}

// ---------- EncryptEntry / DecryptEntry ----------

func TestEncryptDecryptEntry_Success(t *testing.T) {
	key := bytes.Repeat([]byte{7}, 32)
	in := sample{ID: 42, Name: "Alice"}

	ct, nonce, err := EncryptEntry(in, key)
	if err != nil {
		t.Fatalf("EncryptEntry error: %v", err)
	}
	if len(nonce) != 12 {
		t.Fatalf("expected GCM nonce size 12, got %d", len(nonce))
	}
	var out sample
	if err := DecryptEntry(ct, nonce, key, &out); err != nil {
		t.Fatalf("DecryptEntry error: %v", err)
	}
	if out != in {
		t.Fatalf("roundtrip mismatch: got %+v, want %+v", out, in)
	}
}

func TestEncryptEntry_InvalidKeyLength(t *testing.T) {
	key := []byte("tooshort") // 8 bytes
	_, _, err := EncryptEntry(sample{}, key)
	if err == nil {
		t.Fatalf("expected error for invalid AES key length")
	}
}

func TestDecryptEntry_WrongKeyReturnsError(t *testing.T) {
	key := bytes.Repeat([]byte{3}, 32)
	badKey := bytes.Repeat([]byte{4}, 32)
	ct, nonce, err := EncryptEntry(sample{ID: 1}, key)
	if err != nil {
		t.Fatalf("EncryptEntry err: %v", err)
	}
	var out sample
	err = DecryptEntry(ct, nonce, badKey, &out)
	if err == nil {
		t.Fatalf("expected auth error with wrong key")
	}
}

func TestDecryptEntry_TamperedCiphertextReturnsError(t *testing.T) {
	key := bytes.Repeat([]byte{5}, 32)
	ct, nonce, err := EncryptEntry(sample{Name: "X"}, key)
	if err != nil {
		t.Fatalf("EncryptEntry err: %v", err)
	}
	ct[0] ^= 0xFF
	var out sample
	if err := DecryptEntry(ct, nonce, key, &out); err == nil {
		t.Fatalf("expected error for tampered ciphertext")
	}
}

// ---------- EncryptFile / DecryptFile / DecryptFileTo ----------

func TestEncryptDecryptFile_FullRoundtrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "plain.txt")
	want := []byte("hello secret file")
	if err := os.WriteFile(src, want, 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	ef, err := EncryptFile(src)
	if err != nil {
		t.Fatalf("EncryptFile err: %v", err)
	}
	if len(ef.Key) != 32 {
		t.Fatalf("expected file key length 32, got %d", len(ef.Key))
	}
	if len(ef.Nonce) == 0 || len(ef.Cyphertext) == 0 {
		t.Fatalf("nonce/ciphertext should be non-empty")
	}

	got, err := DecryptFile(ef)
	if err != nil {
		t.Fatalf("DecryptFile err: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("file roundtrip mismatch: got %q, want %q", string(got), string(want))
	}

	dst := filepath.Join(dir, "out.txt")
	if err := DecryptFileTo(dst, ef); err != nil {
		t.Fatalf("DecryptFileTo err: %v", err)
	}
	read, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !bytes.Equal(read, want) {
		t.Fatalf("written file mismatch")
	}
}

func TestEncryptFile_MissingPathReturnsError(t *testing.T) {
	_, err := EncryptFile("no/such/file.bin")
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestDecryptFile_Errors(t *testing.T) {
	// nil ef
	if _, err := DecryptFile(nil); err == nil {
		t.Fatalf("expected error for nil EncryptedFile")
	}
	// invalid key length
	ef := &EncryptedFile{Cyphertext: []byte{1}, Key: []byte("short"), Nonce: []byte{1, 2, 3}}
	if _, err := DecryptFile(ef); err == nil {
		t.Fatalf("expected error for invalid key length")
	}
	ef = &EncryptedFile{Cyphertext: []byte{1}, Key: bytes.Repeat([]byte{9}, 32), Nonce: []byte{1, 2, 3}}
	if _, err := DecryptFile(ef); err == nil {
		t.Fatalf("expected error for invalid nonce length")
	}
}

func TestDecryptFile_TamperedReturnsError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "data.bin")
	orig := bytes.Repeat([]byte{0xAB}, 256)
	if err := os.WriteFile(src, orig, 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	ef, err := EncryptFile(src)
	if err != nil {
		t.Fatalf("EncryptFile err: %v", err)
	}
	ef.Cyphertext[0] ^= 0xAA
	if _, err := DecryptFile(ef); err == nil {
		t.Fatalf("expected error for tampered ciphertext")
	}
}

func TestDecryptFileTo_PropagatesError(t *testing.T) {
	err := DecryptFileTo(filepath.Join(t.TempDir(), "x"), &EncryptedFile{
		Cyphertext: []byte{1},
		Key:        []byte("bad"),
		Nonce:      []byte{1},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if _, ok := err.(interface{ Error() string }); !ok {
		t.Fatalf("unexpected error type: %T", err)
	}
}

// ---------- extra: sanity for hex encoding example in docs ----------

func TestHexDocExampleSanity(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	ct, nonce, err := EncryptEntry(sample{ID: 1}, key)
	if err != nil {
		t.Fatalf("EncryptEntry err: %v", err)
	}
	_ = hex.EncodeToString(ct)
	_ = hex.EncodeToString(nonce)
}
