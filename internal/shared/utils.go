// Package shared provides utility functions for working with
// random strings and secure memory wiping.
package shared

import (
	"crypto/rand"
	"encoding/hex"
)

// MakeRandHexString generates a random hexadecimal string of the given size.
// The size parameter specifies the number of random bytes to generate before
// encoding them as a hexadecimal string. As a result, the final string length
// will be twice the size (since each byte expands to two hex characters).
//
// Example:
//
//	s, err := MakeRandHexString(16)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(s) // e.g., "9f2d4c3a5e6b1a7d..."
//
// It returns an error if the random number generator fails.
func MakeRandHexString(size int) (string, error) {

	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// WipeByteArray overwrites the contents of the provided byte slice with zeros.
// This is useful for removing sensitive data such as passwords or cryptographic
// keys from memory after use.
//
// If the slice is nil, the function does nothing.
func WipeByteArray(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
}
