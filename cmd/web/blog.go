package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const blogContentDir = "./content/blog"

var validBlogSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	if !validBlogSlug.MatchString(slug) {
		app.notFound(w, r)
		return
	}

	contentPath := filepath.Join(blogContentDir, slug+".tmpl")
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		contentPath,
		"./ui/html/partials/footer.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		if os.IsNotExist(err) {
			app.notFound(w, r)
			return
		}
		app.serverError(w, r, err)
		return
	}

	pageTitle, err := blogMetadata(ts, "page-title")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	pageDescription, err := blogMetadata(ts, "page-description")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	openGraphDescription, err := blogMetadata(ts, "open-graph-description")
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	openGraphImage, err := blogMetadata(ts, "open-graph-image")
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := PageData{
		PageTitle:            pageTitle,
		PageDescription:      pageDescription,
		OpenGraphDescription: openGraphDescription,
		CanonicalURL:         "https://tabirianalytics.com/blog/" + slug,
		OpenGraphImage:       openGraphImage,
		CurrentPage:          "blog",
		CurrentYear:          time.Now().Year(),
	}

	w.Header().Add("Server", "Go")
	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		app.serverError(w, r, err)
	}
}

func blogMetadata(ts *template.Template, name string) (string, error) {
	var value bytes.Buffer
	if err := ts.ExecuteTemplate(&value, name, nil); err != nil {
		return "", fmt.Errorf("rendering blog metadata %q: %w", name, err)
	}

	metadata := strings.TrimSpace(value.String())
	if metadata == "" {
		return "", fmt.Errorf("blog metadata %q is empty", name)
	}

	return metadata, nil
}
