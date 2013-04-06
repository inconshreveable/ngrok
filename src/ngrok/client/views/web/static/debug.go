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

func BodyHtml() []byte     { return ReadFileOrPanic("assets/body.html") }
func PageHtml() []byte     { return ReadFileOrPanic("assets/page.html") }
func HighlightJs() []byte  { return ReadFileOrPanic("assets/highlight.min.js") }
func HighlightCss() []byte { return ReadFileOrPanic("assets/highlight.min.css") }
func BootstrapCss() []byte { return ReadFileOrPanic("assets/bootstrap.min.css") }
func JqueryJs() []byte     { return ReadFileOrPanic("assets/jquery-1.9.1.min.js") }
