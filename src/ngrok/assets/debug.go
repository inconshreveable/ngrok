// +build !release

package assets

import (
	"io/ioutil"
)

func ReadAsset(path string) (b []byte, err error) {
	return ioutil.ReadFile(path)
}
