package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"os"

	"github.com/dmitrijs2005/gophkeeper/internal/common"
	"golang.org/x/crypto/argon2"
)

func MakeVerifier(masterKey []byte) []byte {
	hash := sha256.Sum256(masterKey)
	return hash[:]
}

func DeriveMasterKey(password []byte, salt []byte) []byte {
	x := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
	return x
}

// EncryptEntry serializes the given entry to JSON and encrypts it using AES-GCM.
//
// The key must be a valid AES key length (16, 24, or 32 bytes for AES-128,
// AES-192, or AES-256 respectively). A new random 12-byte nonce is generated
// for each encryption. The ciphertext and nonce are returned separately.
//
// Parameters:
//   - entry: any Go value that can be marshaled to JSON.
//   - key: the AES encryption key.
//
// Returns:
//   - ciphertext: the encrypted JSON data.
//   - nonce: the randomly generated 12-byte nonce.
//   - err: non-nil if serialization or encryption fails.
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	key := make([]byte, 32) // 256-bit key
//	if _, err := rand.Read(key); err != nil {
//	    log.Fatal(err)
//	}
//
//	user := User{ID: 1, Name: "Alice"}
//
//	ciphertext, nonce, err := EncryptEntry(user, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Encrypted data: %x\n", ciphertext)
//	fmt.Printf("Nonce: %x\n", nonce)
func EncryptEntry(entry any, key []byte) (ciphertext, nonce []byte, err error) {

	// serializing JSON
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return nil, nil, err
	}

	// nonce
	nonce = make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	// new cypher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// encrypting
	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// DecryptEntry decrypts the given ciphertext using AES-GCM and unmarshals
// the resulting JSON into the provided value v.
//
// The key must be the same AES key that was used to encrypt the data,
// and the nonce must be the same 12-byte nonce generated during encryption.
//
// Parameters:
//   - ciphertext: the encrypted data produced by EncryptEntry.
//   - nonce: the 12-byte nonce generated during encryption.
//   - key: the AES encryption key (must be 16, 24, or 32 bytes).
//   - v: a pointer to the Go value into which the decrypted JSON will be unmarshaled.
//
// Returns:
//   - error: non-nil if decryption or JSON unmarshaling fails.
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	// Assume ciphertext, nonce, and key were obtained from EncryptEntry
//	var user User
//	err := DecryptEntry(ciphertext, nonce, key, &user)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Decrypted user: %+v\n", user)
func DecryptEntry(ciphertext, nonce, key []byte, v any) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	return json.Unmarshal(plaintext, v)
}

type EncryptedFile struct {
	Cyphertext []byte
	Key        []byte
	Nonce      []byte
}

func EncryptFile(path string) (*EncryptedFile, error) {
	// reading the file
	plaintext, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// random file_key (32 байта)
	key := common.GenerateRandByteArray(32)

	// creating AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// nonce
	nonce := common.GenerateRandByteArray(aesgcm.NonceSize())

	// шифруем
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	return &EncryptedFile{Cyphertext: ciphertext, Key: key, Nonce: nonce}, nil
}
