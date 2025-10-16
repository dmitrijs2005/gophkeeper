package cryptox

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// func TestGenerateRandByteArray(t *testing.T) {
// 	size := 32
// 	data1 := GenerateRandByteArray(size)
// 	data2 := GenerateRandByteArray(size)
// 	assert.NotEqual(t, data1, data2)
// 	assert.Equal(t, len(data1), size)
// 	assert.Equal(t, len(data2), size)
// }

func TestDeriveMasterKey_Deterministic(t *testing.T) {
	password := []byte("secret-password")
	salt := []byte("fixed-salt")

	key1 := DeriveMasterKey(password, salt)
	key2 := DeriveMasterKey(password, salt)

	// одинаковые входы -> одинаковый вывод
	if !bytes.Equal(key1, key2) {
		t.Errorf("expected same result for same inputs, got different")
	}

	// можно зафиксировать известный результат (snapshot test)
	expectedHex := "9290403300158e19f27e48e7087f7383b03065bf5b25ef23ebc40229616cd8b3"
	if hex.EncodeToString(key1) != expectedHex {
		t.Errorf("expected %s, got %s", expectedHex, hex.EncodeToString(key1))
	}
}

func TestDeriveMasterKey_DifferentInputs(t *testing.T) {
	password := []byte("secret-password")
	salt1 := []byte("salt-1")
	salt2 := []byte("salt-2")

	key1 := DeriveMasterKey(password, salt1)
	key2 := DeriveMasterKey(password, salt2)

	if bytes.Equal(key1, key2) {
		t.Errorf("expected different results for different salts, got same")
	}
}

func encryptForTest(t *testing.T, plaintext []byte) *EncryptedFile {
	t.Helper()

	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err)

	block, err := aes.NewCipher(key)
	require.NoError(t, err)

	aesgcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	nonce := make([]byte, aesgcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	require.NoError(t, err)

	ct := aesgcm.Seal(nil, nonce, plaintext, nil)
	return &EncryptedFile{Cyphertext: ct, Key: key, Nonce: nonce}
}

func TestDecryptFile_Success(t *testing.T) {
	t.Parallel()

	plain := []byte("hello, gcm!")
	ef := encryptForTest(t, plain)

	out, err := DecryptFile(ef)
	require.NoError(t, err)
	require.Equal(t, plain, out)
}

func TestDecryptFile_EmptyPlaintext(t *testing.T) {
	t.Parallel()

	plain := []byte{}
	ef := encryptForTest(t, plain)
	out, err := DecryptFile(ef)
	require.NoError(t, err)

	require.Zero(t, len(out), "empty plaintext must decrypt to zero-length slice")
}

func TestDecryptFile_InvalidKeyLength(t *testing.T) {
	t.Parallel()

	ef := encryptForTest(t, []byte("x"))
	ef.Key = ef.Key[:16]

	_, err := DecryptFile(ef)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key length")
}

func TestDecryptFile_InvalidNonceLength(t *testing.T) {
	t.Parallel()

	ef := encryptForTest(t, []byte("x"))
	ef.Nonce = ef.Nonce[:len(ef.Nonce)-1]

	_, err := DecryptFile(ef)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid nonce length")
}

func TestDecryptFile_TamperedCiphertext(t *testing.T) {
	t.Parallel()

	ef := encryptForTest(t, []byte("authenticated data"))

	ef.Cyphertext[0] ^= 0xFF

	_, err := DecryptFile(ef)
	require.Error(t, err, "tampered ciphertext must fail authentication")
}

func TestDecryptFileTo_WritesFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	outPath := filepath.Join(tmp, "out.bin")

	orig := bytes.Repeat([]byte{0xAB}, 1024) // 1 KiB
	ef := encryptForTest(t, orig)

	err := DecryptFileTo(outPath, ef)
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	require.Equal(t, orig, got)
}
