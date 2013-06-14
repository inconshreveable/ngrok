// +build !release

package static

import (
	"io/ioutil"
	"os"
	"path"
)

var assetDir string

func init() {
	// find the directory with the assets.
	// this doesn't work if you:
	// 1. move the binary
	// 2. put ngrok in your PATH
	// but you shouldn't be doing either of these things while developng anyways
	var binPath string
	execPath := os.Args[0]
	if path.IsAbs(execPath) {
		binPath = execPath
	} else {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		binPath = path.Join(wd, execPath)
	}
	assetDir = path.Join(path.Dir(path.Dir(binPath)), "assets")

	// call all the functions on startup to make sure the files exist
	fns := []func() []byte{
		PageHtml,
		HighlightJs,
		HighlightCss,
		BootstrapCss,
		JqueryJs,
		VkBeautifyJs,
		AngularJs,
		NgrokJs,
		Base64Js,
	}
	for _, f := range fns {
		f()
	}
}

func ReadFileOrPanic(p string) []byte {
	bytes, err := ioutil.ReadFile(path.Join(assetDir, p))
	if err != nil {
		panic(err)
	}
	return bytes
}

func PageHtml() []byte     { return ReadFileOrPanic("page.html") }
func HighlightJs() []byte  { return ReadFileOrPanic("highlight.min.js") }
func HighlightCss() []byte { return ReadFileOrPanic("highlight.min.css") }
func BootstrapCss() []byte { return ReadFileOrPanic("bootstrap.min.css") }
func JqueryJs() []byte     { return ReadFileOrPanic("jquery-1.9.1.min.js") }
func VkBeautifyJs() []byte { return ReadFileOrPanic("vkbeautify.js") }
func AngularJs() []byte    { return ReadFileOrPanic("angular.js") }
func NgrokJs() []byte      { return ReadFileOrPanic("ngrok.js") }
func Base64Js() []byte     { return ReadFileOrPanic("base64.js") }
