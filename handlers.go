package httpex

import (
	"io/fs"
	"log"
	"net/http"
)

// Middleware is a function that wraps an http.Handler and returns a new
// http.Handler.
type Middleware func(next http.Handler) http.Handler

// FileServerSubdir returns a file server from a subdir of the given file
// system, with the given url path. If urlpath is empty, it will use "/".  If
// subdir is empty, it will use the root of the file system. It will panic if it
// fails to initialise a subfs.
func FileServerSubdir(fsys fs.FS, subdir string, urlpath string) http.Handler {
	if subdir == "" {
		return FileServer(fsys, urlpath)
	}
	subfs, err := fs.Sub(fsys, subdir)
	if err != nil {
		log.Panicf("internal error: fs.Sub: %v", err)
	}
	return FileServer(subfs, urlpath)
}

// FileServer returns a file server from the given file system, with the given
// url path.  If urlpath is empty, it will use "/".
func FileServer(fs fs.FS, urlpath string) http.Handler {
	if urlpath == "" {
		urlpath = "/"
	}
	return Neuter(urlpath, http.StripPrefix(urlpath, http.FileServer(http.FS(fs))))
}
