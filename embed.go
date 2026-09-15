package web

import (
	"embed"
	"io/fs"
)

//go:embed static/index.html static/assets
var files embed.FS

var Index []byte

func init() {
	var err error
	Index, err = files.ReadFile("static/index.html")
	if err != nil {
		panic(err)
	}
}

func Assets() fs.FS {
	assets, err := fs.Sub(files, "static/assets")
	if err != nil {
		panic(err)
	}
	return assets
}
