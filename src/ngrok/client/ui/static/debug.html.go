// +build !release

package static

import (
    "io/ioutil"
)

func ReadFileOrPanic(path string) []byte {
    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        panic(err)
    }
    return bytes
}

func BodyHtml() []byte { return ReadFileOrPanic("templates/body.html") }
func PageHtml() []byte { return ReadFileOrPanic("templates/page.html") }
