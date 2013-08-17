package util

import (
	"crypto/rand"
	"fmt"
)

func RandomSeed() (int64, error) {
	b := make([]byte, 8)
	n, err := rand.Read(b)
	if n != 8 {
		return 0, fmt.Errorf("Only generated %d random bytes, %d requested", n, 8)
	}

	if err != nil {
		return 0, err
	}

	var seed int64
	var i uint
	for i = 0; i < 8; i++ {
		seed = seed | int64(b[i]<<(i*8))
	}

	return seed, nil
}

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
