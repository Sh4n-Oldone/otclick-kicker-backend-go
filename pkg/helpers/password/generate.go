package password

import (
	"math/rand/v2"
)

func Generate() string {
	symbols := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()+-=[]{}|;':,./<>?")

	idx := len(symbols)

	result := make([]byte, 16)
	for i, _ := range result {
		result[i] = symbols[rand.IntN(idx)]
	}

	return string(result)
}
