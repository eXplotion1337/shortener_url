package internal

import (
	mr "math/rand"
	"time"
)

const charset = "0123456789"

var seededRand *mr.Rand = mr.New(mr.NewSource(time.Now().UnixNano()))

func GenerateRandomString(length int) string {
	result := make([]byte, length)
	// charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n := seededRand.Int63n(int64(len(charset)))
		result[i] = charset[n]
	}

	return string(result)
}
