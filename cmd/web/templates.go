package main

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

const (
	baseTemplate   = "./ui/html/base.tmpl"
	navTemplate    = "./ui/html/partials/nav.tmpl"
	footerTemplate = "./ui/html/partials/footer.tmpl"
	pagesDir       = "./ui/html/pages"
)

type templateCache struct {
	pages map[string]*template.Template
	blogs map[string]*template.Template
}

func newTemplateCache() (*templateCache, error) {
	cache := &templateCache{
		pages: make(map[string]*template.Template, len(pageSpecs)),
		blogs: make(map[string]*template.Template),
	}

	for _, spec := range pageSpecs {
		if _, exists := cache.pages[spec.template]; exists {
			continue
		}
		path := filepath.Join(pagesDir, spec.template+".tmpl")
		tmpl, err := template.ParseFiles(baseTemplate, navTemplate, path, footerTemplate)
		if err != nil {
			return nil, fmt.Errorf("parse page template %q: %w", spec.template, err)
		}
		cache.pages[spec.template] = tmpl
	}

	blogPaths, err := filepath.Glob(filepath.Join(blogContentDir, "*.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("list blog templates: %w", err)
	}
	for _, path := range blogPaths {
		slug := trimTemplateExtension(filepath.Base(path))
		tmpl, err := template.ParseFiles(baseTemplate, navTemplate, path, footerTemplate)
		if err != nil {
			return nil, fmt.Errorf("parse blog template %q: %w", slug, err)
		}
		cache.blogs[slug] = tmpl
	}
	return cache, nil
}

func trimTemplateExtension(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func executeTemplate(tmpl *template.Template, data PageData) ([]byte, error) {
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "base", data); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
