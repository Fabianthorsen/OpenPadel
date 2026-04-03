package ui

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// Files holds the compiled SvelteKit output.
// The `build` directory is populated by `bun run build` in web/ before `go build`.
//
//go:embed all:build
var files embed.FS

// Handler returns an http.Handler that serves the SvelteKit SPA.
// Static assets are served directly; everything else falls back to index.html.
func Handler() http.Handler {
	sub, err := fs.Sub(files, "build")
	if err != nil {
		panic("ui: failed to sub embedded files: " + err.Error())
	}

	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		// No frontend build present — serve a placeholder during dev
		log.Println("ui: no index.html found in embedded build, serving placeholder")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("frontend not built — run `make build`"))
		})
	}

	log.Printf("ui: serving embedded frontend (%d byte index.html)", len(index))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to serve the asset directly (JS, CSS, images, etc.)
		// Serve using ServeContent to avoid any redirect behavior from http.FileServer.
		if path != "" && path != "index.html" {
			f, err := sub.Open(path)
			if err == nil {
				stat, serr := f.Stat()
				if serr == nil && !stat.IsDir() {
					log.Printf("ui: static %s", r.URL.Path)
					ctype := mime.TypeByExtension(filepath.Ext(path))
					if ctype != "" {
						w.Header().Set("Content-Type", ctype)
					}
					http.ServeContent(w, r, path, stat.ModTime(), f.(io.ReadSeeker))
					f.Close()
					return
				}
				f.Close()
			}
		}

		// SPA fallback — serve index.html for all unmatched routes
		log.Printf("ui: spa fallback for %s", r.URL.Path)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index)
	})
}
