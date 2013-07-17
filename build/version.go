package main

import (
	"fmt"
	"ngrok/version"
)

func main() {
	fmt.Print(version.MajorMinor())
}
