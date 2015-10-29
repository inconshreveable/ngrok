package util

import (
	"time"
	mrand "math/rand"
)

// Generate n random characters for use in URLs
func RandomChars(n int) string {
  mrand.Seed(time.Now().UnixNano())
  var chars = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
  bytes := make([]rune, n)
  for i := range bytes {
    bytes[i] = chars[mrand.Intn(len(chars))]
  }
  return string(bytes)
}

