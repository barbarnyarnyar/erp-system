package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// NewID generates a randomized and timestamped ID with a given prefix to avoid collision.
func NewID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s_%d", prefix, hex.EncodeToString(b), time.Now().UnixNano())
}
