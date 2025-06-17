package random

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	mathRand "math/rand"
	"time"

	"github.com/google/uuid"
)

var Rand *mathRand.Rand

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// String generates a random string of the specified length
func String(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple method if crypto/rand fails
		fallbackBytes := make([]byte, length/2)
		for i := range fallbackBytes {
			fallbackBytes[i] = byte(i % 256)
		}
		return hex.EncodeToString(fallbackBytes)
	}
	return hex.EncodeToString(bytes)
}

func Token() string {
	return "openlist-" + uuid.NewString() + String(64)
}

func RangeInt64(left, right int64) int64 {
	return mathRand.Int63n(left+right) - left
}

func init() {
	s := mathRand.NewSource(time.Now().UnixNano())
	Rand = mathRand.New(s)
}
