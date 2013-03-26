package util

import (
	"crypto/rand"
	"fmt"
)

// create a random identifier for this client
func RandId() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("rand.Rand failed trying to create identifier, %v", err))
	}
	return fmt.Sprintf("%x", b)
}
