package assets

import (
	"embed"
	"strings"
)

//go:embed client/*
//go:embed server/*
var assetsFS embed.FS

func Asset(path string) ([]byte, error) {
	return assetsFS.ReadFile(strings.TrimPrefix(path, "assets/"))
}
