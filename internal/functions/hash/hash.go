// Package hash provides functions
// working with hashing.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
)

// MakeHashSHA256 - caches data with key.
func MakeHashSHA256(dataMsg *[]byte,
	key string,
) ([]byte, error) {
	dataKey := []byte(key)

	h := hmac.New(sha256.New, dataKey)
	h.Write(*dataMsg)
	sign := h.Sum(nil)

	return sign, nil
}
