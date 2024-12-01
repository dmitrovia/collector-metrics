package hash

import (
	"crypto/hmac"
	"crypto/sha256"
)

func MakeHashSHA256(dataMsg *[]byte,
	key string,
) ([]byte, error) {
	dataKey := []byte(key)

	h := hmac.New(sha256.New, dataKey)
	h.Write(*dataMsg)
	sign := h.Sum(nil)

	return sign, nil
}
