package util

import (
	"math/rand"
	"time"
)

const (
	RandomStringPool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// RandomStringGenerator generates a random string of fixed length
func RandomStringGenerator(length int) string {
	if length <= 0 {
		return ""
	}

	pool_rune := []rune(RandomStringPool)
	rand.Seed(time.Now().UnixNano())
	result := make([]rune, length)

	for i := range result {
		result[i] = pool_rune[rand.Intn(len(pool_rune))]
	}
	return string(result)
}
