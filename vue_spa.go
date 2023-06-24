package httpex

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewVueSPA returns a new VueSPA instance.  fsys is the file system to serve
// files from.  rootPath is the path to the root Vue compiled "dist" directory
// within the fsys, and staticDirName is the name of the "assets" directory
// within the rootPath, that contains all *.js and *.css files.
//
// For example, consider this file system:
//
//	.
//	└── frontend
//	    └── dist
//	        ├── favicon.ico
//	        ├── index.html
//	        └── assets
//	            ├── css
//	            │   ├── app.0a0b1c2d.css
//	            │   └── chunk-vendors.3e4f5g6h.css
//	            └── js
//	                ├── app.0a0b1c2d.js
//	                └── chunk-vendors.3e4f5g6h.js
//
// Then, the following code will serve the Vue SPA:
//
//	mux, err := httpex.NewVueSPA(fsys, "frontend/dist", "assets")
//	if err != nil {
//		log.Fatal(err)
//	}
func NewVueSPA(fsys fs.FS, rootPath, assetsDirName string) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	rootfs, err := fs.Sub(fsys, rootPath)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(assetsDirName, "/") {
		assetsDirName += "/"
	}

	// Serve static files from the embedded file system.
	mux.Handle("/"+assetsDirName, Neuter("/", http.FileServer(http.FS(rootfs))))
	mux.Handle("/", vueIndexHandler(rootfs))

	return mux, nil
}

// vueIndexHandler serves favicon.ico if favicon is requested, or index.html in
// any other case from the embedded filesystem.  This is a helper function for
// NewVueSPA.
//
// It uses the feature of the pattern matching in the http.ServeMux, that
// forwards the request to the handler registered with the longest matching
// pattern.  Such as the request for anything other than favicon.ico will be
// forwarded to the Vue router via serving the Vue's index html.
//
// Example:
//
//	Request to: /favicon.ico  => Serve favicon.ico
//	Request to: /             => Serve index.html
//	Request to: /foo          => Serve index.html, and if Vue router has a
//	                             route for /foo, then it will be handled by
//	                             Vue router.
func vueIndexHandler(fsys fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			content, err := fs.ReadFile(fsys, "favicon.ico")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(content)
			return
		}
		content, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(content)
	}
}
