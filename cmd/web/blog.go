package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

const blogContentDir = "./content/blog"

var validBlogSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (app *application) blogPost(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("id")
	if !validBlogSlug.MatchString(slug) {
		app.notFound(w, r)
		return
	}

	ts, ok := app.templates.blogs[slug]
	if !ok {
		app.notFound(w, r)
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
		CurrentYear:          app.now().Year(),
	}

	output, err := executeTemplate(ts, data)
	if err != nil {
		app.serverError(w, r, fmt.Errorf("render blog template %q: %w", slug, err))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(output); err != nil {
		app.logger.ErrorContext(r.Context(), "write response", "error", err)
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
