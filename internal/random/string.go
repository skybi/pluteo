package random

import "math/rand"

var (
	// CharsetAlphanumeric contains characters a-zA-Z0-9
	CharsetAlphanumeric = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890")

	// CharsetTokens contains all alphanumeric characters plus '.-#+*~'
	CharsetTokens = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890.-#+*~")
)

// String generates a random string with a specific length, only using characters out of the given charset
func String(length int, charset []rune) string {
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = charset[rand.Intn(len(charset))]
	}
	return string(buf)
}
