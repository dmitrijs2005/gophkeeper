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
	expectedHex := "34f7a1c64df63ab1ad5b5ee06e64db5713b35f81839823304db63e8e5e6a6a39"
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
