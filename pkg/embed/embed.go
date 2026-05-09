package embed

import (
	"embed"
	"io/fs"
)

//go:embed all:public
var publicFS embed.FS

func GetPublicFS() fs.FS {
	sub, _ := fs.Sub(publicFS, "public")
	return sub
}

func GetDocsFS() fs.FS {
	return nil
}
