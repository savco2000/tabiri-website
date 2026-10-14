package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlogPostTemplatesParse(t *testing.T) {
	t.Chdir("../..")

	for _, slug := range blogPostSlugs(t) {
		t.Run(slug, func(t *testing.T) {
			_, err := template.ParseFiles(
				"./ui/html/base.tmpl",
				"./ui/html/partials/nav.tmpl",
				filepath.Join(blogContentDir, slug+".tmpl"),
				"./ui/html/partials/footer.tmpl",
			)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBlogPostRendersForEverySlug(t *testing.T) {
	app, _ := newTestApplication(t)

	for _, slug := range blogPostSlugs(t) {
		t.Run(slug, func(t *testing.T) {
			response := getPage(app, "/blog/"+slug)

			if response.Code != http.StatusOK {
				t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
			}

			body := response.Body.String()

			canonical := "https://tabirianalytics.com/blog/" + slug
			if !strings.Contains(body, canonical) {
				t.Errorf("response does not contain canonical URL %q", canonical)
			}
			if !strings.Contains(body, `<meta property="og:image" content="https://`) {
				t.Error("response does not contain an absolute Open Graph image")
			}
			// Relative asset paths resolve against /blog/ and would 404.
			if strings.Contains(body, `src="./`) {
				t.Error("response contains a relative asset path")
			}
			if strings.Contains(body, ".html\"") {
				t.Error("response contains a stale .html link")
			}
		})
	}
}

func TestBlogListingLinksToEveryPost(t *testing.T) {
	app, _ := newTestApplication(t)

	response := getPage(app, "/blog")
	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	for _, slug := range blogPostSlugs(t) {
		link := "href=\"/blog/" + slug + "\""
		if !strings.Contains(body, link) {
			t.Errorf("blog listing does not link to %q", link)
		}
	}
}

func TestBlogPostUnknownSlugReturnsNotFound(t *testing.T) {
	app, _ := newTestApplication(t)

	response := getPage(app, "/blog/no-such-post")

	if response.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusNotFound)
	}
}

func getPage(app *application, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)
	return response
}

func blogPostSlugs(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(blogContentDir, "*.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no blog content templates found")
	}

	slugs := make([]string, 0, len(paths))
	for _, path := range paths {
		slugs = append(slugs, strings.TrimSuffix(filepath.Base(path), ".tmpl"))
	}
	return slugs
}
