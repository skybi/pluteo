package secret

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// MustNew generates a new base64-encoded secret + its base64-encoded SHA256 hash
func MustNew(len int) (string, string) {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	raw := base64.StdEncoding.EncodeToString(bytes)
	sum := sha256.Sum256(bytes)
	return raw, base64.StdEncoding.EncodeToString(sum[:])
}

// Hash decodes the given base64 string and returns its base64-encoded SHA256 hash
func Hash(raw string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(bytes)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}
