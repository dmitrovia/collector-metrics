package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func Intn(maximum int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maximum))
	if err != nil {
		return 0, fmt.Errorf("Intn->rand.Int: %w", err)
	}

	return nBig.Int64(), nil
}

func RandF64(maximum int64) (float64, error) {
	const shift = 53

	value, err := Intn(maximum << shift)
	if err != nil {
		return 0, fmt.Errorf("Intn->rand.Int: %w", err)
	}

	return float64(value) / (1 << shift), nil
}
