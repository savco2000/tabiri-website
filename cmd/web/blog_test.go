package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlogPostTemplatesParse(t *testing.T) {
	t.Chdir("../..")

	for slug, post := range blogPosts {
		t.Run(slug, func(t *testing.T) {
			_, err := template.ParseFiles(
				"./ui/html/base.tmpl",
				"./ui/html/partials/nav.tmpl",
				post.Template,
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

	for slug, post := range blogPosts {
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
			if !strings.Contains(body, post.OpenGraphImage) {
				t.Errorf("response does not contain Open Graph image %q", post.OpenGraphImage)
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
	for slug := range blogPosts {
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
