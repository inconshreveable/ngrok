package util

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
)

func RandomSeed() (seed int64, err error) {
	err = binary.Read(rand.Reader, binary.LittleEndian, &seed)
	return
}

// creates a random identifier of the specified length
func RandId(idlen int) string {
	b := make([]byte, idlen)
	var randVal uint32
	for i := 0; i < idlen; i++ {
		byteIdx := i % 4
		if byteIdx == 0 {
			randVal = mrand.Uint32()
		}
		b[i] = byte((randVal >> (8 * uint(byteIdx))) & 0xFF)
	}
	return fmt.Sprintf("%x", b)
}

// like RandId, but uses a crypto/rand for secure random identifiers
func SecureRandId(idlen int) (id string, err error) {
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

func SecureRandIdOrPanic(idlen int) string {
	id, err := SecureRandId(idlen)
	if err != nil {
		panic(err)
	}
	return id
}
