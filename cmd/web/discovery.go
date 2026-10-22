package main

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"time"
)

//go:embed public/robots.txt public/sitemap.xml public/.well-known/security.txt public/healthz
var discoveryFiles embed.FS

func (app *application) discoveryFile(name, contentType, cacheControl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := discoveryFiles.ReadFile("public/" + name)
		if err != nil {
			app.serverError(w, r, fmt.Errorf("read discovery file %q: %w", name, err))
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", cacheControl)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(content))
	})
}
