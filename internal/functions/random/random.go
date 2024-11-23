package random

import (
	"crypto/rand"
	"math/big"
)

func Intn(maximum int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maximum))
	if err != nil {
		panic(err)
	}

	return nBig.Int64()
}

func RandF64(maximum int64) float64 {
	const shift = 53

	return float64(Intn(maximum<<shift)) / (1 << shift)
}
