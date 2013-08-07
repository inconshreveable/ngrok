package util

import (
	"crypto/rand"
	"fmt"
)

// create a random identifier for this client
func RandId(idlen int) (id string, err error) {
	b := make([]byte, idlen)
	n, err := rand.Read(b)

	if n != idlen {
		err = fmt.Errorf("Only generated %d random bytes, %d requested", n, idlen)
		return
	}

	if err != nil {
		return
	}

	id = fmt.Sprintf("%x", b)
	return
}

func RandIdOrPanic(idlen int) string {
	id, err := RandId(idlen)
	if err != nil {
		panic(err)
	}
	return id
}
