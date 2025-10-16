package cryptox

import (
	"bytes"
	"encoding/hex"
	"testing"
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

	// разные соли должны дать разные ключи
	if bytes.Equal(key1, key2) {
		t.Errorf("expected different results for different salts, got same")
	}
}
