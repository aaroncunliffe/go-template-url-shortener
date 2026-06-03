// Very simple generation package using sensible characters to support the required entropy

package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const defaultLength = 6
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate() (string, error) {
	return GenerateLength(defaultLength)
}

func GenerateLength(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than zero")

	}

	code := make([]byte, length)
	for i := range code {
		// Slightly slower crypto/rand, but provides a much more reliable random
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}
