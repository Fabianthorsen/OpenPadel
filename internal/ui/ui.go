package ui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// Files holds the compiled SvelteKit output.
// The `build` directory is populated by `bun run build` in web/ before `go build`.
//
//go:embed build
var files embed.FS

// Handler returns an http.Handler that serves the SvelteKit SPA.
// Any path not found falls back to index.html for client-side routing.
func Handler() http.Handler {
	sub, err := fs.Sub(files, "build")
	if err != nil {
		panic("ui: failed to sub embedded files: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		// If the file exists, serve it directly.
		if f, err := sub.Open(path); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Otherwise fall back to index.html for SPA routing.
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r2)
	})
}
