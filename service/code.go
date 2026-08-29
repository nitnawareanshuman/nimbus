package service

import (
	"crypto/rand"
	"math/big"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateCode() string {
	const length = 6

	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader,big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}

		result[i] = charset[n.Int64()]
	}

	return string(result)
}