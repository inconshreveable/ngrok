// +build !release

package assets

import (
	"io/ioutil"
)

func ReadAsset(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}
