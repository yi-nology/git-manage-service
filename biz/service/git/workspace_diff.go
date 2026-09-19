package git

import (
	"path/filepath"
	"strings"
)

type diffOp struct {
	kind byte
	line string
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
		".ico": true, ".svg": true, ".pdf": true, ".zip": true, ".tar": true,
		".gz": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".mp3": true,
		".mp4": true, ".avi": true, ".mov": true, ".wasm": true,
	}
	return binaryExts[ext]
}
