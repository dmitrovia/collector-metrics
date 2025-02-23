// Package random provides functions
// working with random values.
package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Intn - gets big int64 value.
func Intn(maximum int64) (int64, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maximum))
	if err != nil {
		return 0, fmt.Errorf("Intn->rand.Int: %w", err)
	}

	return nBig.Int64(), nil
}

// RandF64 - gets random float64 value.
func RandF64(maximum int64) (float64, error) {
	const shift = 53

	value, err := Intn(maximum << shift)
	if err != nil {
		return 0, fmt.Errorf("Intn->rand.Int: %w", err)
	}

	return float64(value) / (1 << shift), nil
}
