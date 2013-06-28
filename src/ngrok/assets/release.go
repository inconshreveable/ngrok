// +build release

package assets

import (
	"bitbucket.org/tebeka/nrsc"
	"fmt"
	"io/ioutil"
)

func init() {
	nrsc.Initialize()
}

func ReadAsset(path string) (b []byte, err error) {
	resource := nrsc.Get(path)
	if resource == nil {
		err = fmt.Errorf("Asset %s not compiled into package", path)
		return
	}

	rd, err := resource.Open()
	if err != nil {
		return
	}

	return ioutil.ReadAll(rd)
}
