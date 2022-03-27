package secret

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
)

// MustNew generates a new cryptographically secure byte array of length len and returns its base64 representation + its SHA512 hash
func MustNew(len int) (string, [64]byte) {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	raw := base64.StdEncoding.EncodeToString(bytes)
	sum := sha512.Sum512(bytes)
	return raw, sum
}

// Hash decodes the given base64 string and returns its SHA512 hash
func Hash(raw string) ([64]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return [64]byte{}, err
	}
	sum := sha512.Sum512(bytes)
	return sum, nil
}
