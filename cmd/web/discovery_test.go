package main

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryEndpoints(t *testing.T) {
	app, _ := newTestApplication(t)
	for _, test := range []struct {
		path        string
		contentType string
		content     string
		cache       string
	}{
		{"/robots.txt", "text/plain; charset=utf-8", "Sitemap: https://tabirianalytics.com/sitemap.xml", "public, max-age=3600"},
		{"/sitemap.xml", "application/xml; charset=utf-8", "<urlset", "public, max-age=3600"},
		{"/.well-known/security.txt", "text/plain; charset=utf-8", "Contact: mailto:savannah.kadima@tabirianalytics.com", "public, max-age=3600"},
		{"/healthz", "text/plain; charset=utf-8", "ok\n", "no-store"},
	} {
		t.Run(test.path, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
				response := httptest.NewRecorder()
				app.routes().ServeHTTP(response, httptest.NewRequest(method, test.path, nil))
				if method == http.MethodPost {
					if response.Code != http.StatusMethodNotAllowed {
						t.Errorf("POST: got status %d, want 405", response.Code)
					}
					continue
				}
				if response.Code != http.StatusOK {
					t.Fatalf("%s: got status %d, want 200", method, response.Code)
				}
				if got := response.Header().Get("Content-Type"); got != test.contentType {
					t.Errorf("%s: got content type %q", method, got)
				}
				if got := response.Header().Get("Cache-Control"); got != test.cache {
					t.Errorf("%s: got cache control %q", method, got)
				}
				if method == http.MethodHead && response.Body.Len() != 0 {
					t.Error("HEAD response has a body")
				}
				if method == http.MethodGet && !strings.Contains(response.Body.String(), test.content) {
					t.Errorf("GET response missing %q", test.content)
				}
			}
		})
	}
}

func TestSitemapCoversPublishedPagesAndBlogPosts(t *testing.T) {
	app, _ := newTestApplication(t)
	var sitemap struct {
		XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
		URLs    []struct {
			Location string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(getPage(app, "/sitemap.xml").Body.Bytes(), &sitemap); err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{}
	for name, spec := range pageSpecs {
		switch name {
		case "not-found", "privacy-policy", "terms-of-service", "security-disclosure":
			// Unapproved legal drafts are deliberately excluded.
			continue
		}
		location := spec.canonicalURL
		if location == "" {
			location = "https://tabirianalytics.com/" + name
		}
		expected[location] = true
	}
	for slug := range app.templates.blogs {
		expected["https://tabirianalytics.com/blog/"+slug] = true
	}
	seen := map[string]bool{}
	for _, entry := range sitemap.URLs {
		location := entry.Location
		if !expected[location] {
			t.Errorf("unexpected sitemap URL %q", location)
		}
		if seen[location] {
			t.Errorf("duplicate sitemap URL %q", location)
		}
		seen[location] = true
		path := strings.TrimPrefix(location, "https://tabirianalytics.com")
		response := getPage(app, path)
		if response.Code != http.StatusOK {
			t.Errorf("%s returned %d", location, response.Code)
		}
		if !strings.Contains(response.Body.String(), `rel="canonical" href="`+location+`"`) {
			t.Errorf("%s does not match its page canonical URL", location)
		}
	}
	for location := range expected {
		if !seen[location] {
			t.Errorf("sitemap missing %s; update the static sitemap when adding pages", location)
		}
	}
}

func TestSecurityTXTExpiryAndCanonical(t *testing.T) {
	app, _ := newTestApplication(t)
	body := getPage(app, "/.well-known/security.txt").Body.String()
	if !strings.Contains(body, "Canonical: https://tabirianalytics.com/.well-known/security.txt\n") {
		t.Error("missing canonical security.txt URL")
	}
	var expiry time.Time
	for _, line := range strings.Split(body, "\n") {
		if value, ok := strings.CutPrefix(line, "Expires: "); ok {
			var err error
			expiry, err = time.Parse(time.RFC3339, value)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if !expiry.After(time.Now()) {
		t.Fatal("security.txt is missing an expiry or has expired; renew its Expires field")
	}
}

func TestFooterDoesNotClaimSystemStatus(t *testing.T) {
	app, _ := newTestApplication(t)
	body := getPage(app, "/").Body.String()
	if strings.Contains(body, "All Systems Operational") {
		t.Error("footer still claims system status")
	}
	if !strings.Contains(body, `id="cookie-settings"`) {
		t.Error("footer no longer provides cookie settings")
	}
}
