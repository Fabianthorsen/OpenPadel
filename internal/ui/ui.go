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
// Static assets are served directly; everything else falls back to index.html.
func Handler() http.Handler {
	sub, err := fs.Sub(files, "build")
	if err != nil {
		panic("ui: failed to sub embedded files: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		// No frontend build present — serve a placeholder during dev
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("frontend not built — run `make build`"))
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to serve the asset directly (JS, CSS, images, etc.)
		if path != "" && path != "index.html" {
			if f, err := sub.Open(path); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA fallback — serve index.html directly to avoid redirect loops
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})
}
