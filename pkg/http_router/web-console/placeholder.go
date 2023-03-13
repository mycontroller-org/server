// +build !web

package web

import (
	"embed"
	"net/http"
)

//go:embed placeholder/index.html
var assetsFS embed.FS

// StaticFiles provides http filesystem with static files for UI
var StaticFiles = &httpFSWrapper{prefix: "placeholder", fs: http.FS(assetsFS)}

// http.FS lists files under "placeholder/"
// to get it under "/" add a wrapper
type httpFSWrapper struct {
	prefix string
	fs     http.FileSystem
}

func (fs *httpFSWrapper) Open(name string) (http.File, error) {
	prefixedName := fs.prefix + name
	if name == "/" {
		prefixedName = fs.prefix
	}
	return fs.fs.Open(prefixedName)
}
