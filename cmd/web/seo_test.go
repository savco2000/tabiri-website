package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPublicPagesHavePageSpecificCanonicalURLs(t *testing.T) {
	app, _ := newTestApplication(t)
	tests := []struct {
		path      string
		canonical string
	}{
		{"/", "https://tabirianalytics.com/"},
		{"/digital-employees", "https://tabirianalytics.com/digital-employees"},
		{"/operational-assessment", "https://tabirianalytics.com/operational-assessment"},
		{"/penetration-testing", "https://tabirianalytics.com/penetration-testing"},
		{"/managed-cybersecurity-services", "https://tabirianalytics.com/managed-cybersecurity-services"},
		{"/cybersecurity-insurance", "https://tabirianalytics.com/cybersecurity-insurance"},
		{"/data-protection", "https://tabirianalytics.com/data-protection"},
		{"/case-studies", "https://tabirianalytics.com/case-studies"},
		{"/channel-partners", "https://tabirianalytics.com/channel-partners"},
		{"/capability-statement", "https://tabirianalytics.com/capability-statement"},
		{"/blog", "https://tabirianalytics.com/blog"},
		{"/about", "https://tabirianalytics.com/about"},
		{"/contact", "https://tabirianalytics.com/contact"},
		{"/partner-contact", "https://tabirianalytics.com/partner-contact"},
		{"/newsletter", "https://tabirianalytics.com/newsletter"},
		{"/automation-assessment", "https://tabirianalytics.com/automation-assessment"},
		{"/privacy-policy", "https://tabirianalytics.com/privacy-policy"},
		{"/terms-of-service", "https://tabirianalytics.com/terms-of-service"},
		{"/security-disclosure", "https://tabirianalytics.com/security-disclosure"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			response := getPage(app, test.path)
			if response.Code != http.StatusOK {
				t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
			}

			canonicalTag := fmt.Sprintf(`<link rel="canonical" href="%s">`, test.canonical)
			openGraphTag := fmt.Sprintf(`<meta property="og:url" content="%s">`, test.canonical)
			if strings.Count(response.Body.String(), canonicalTag) != 1 {
				t.Errorf("response must contain exactly one canonical tag for %s", test.canonical)
			}
			if strings.Count(response.Body.String(), openGraphTag) != 1 {
				t.Errorf("response must contain exactly one Open Graph URL for %s", test.canonical)
			}
		})
	}
}

func TestHomepageHasRelevantSocialMetadata(t *testing.T) {
	app, _ := newTestApplication(t)
	response := getPage(app, "/")
	body := response.Body.String()

	title := "Secure AI-Powered Digital Employees | Tabiri Analytics"
	if !strings.Contains(body, `<meta property="og:title" content="`+title+`">`) {
		t.Errorf("homepage does not contain the expected Open Graph title")
	}
	if !strings.Contains(body, `<meta name="twitter:title" content="`+title+`">`) {
		t.Errorf("homepage does not contain the expected Twitter title")
	}
}

func TestNotFoundPageDoesNotUseHomepageAsCanonical(t *testing.T) {
	app, _ := newTestApplication(t)
	response := getPage(app, "/missing-page")

	if response.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusNotFound)
	}
	if strings.Contains(response.Body.String(), `rel="canonical"`) ||
		strings.Contains(response.Body.String(), `property="og:url"`) {
		t.Error("not-found page should not claim the homepage as its canonical URL")
	}
}

func TestFooterLegalLinksRenderDraftPages(t *testing.T) {
	app, _ := newTestApplication(t)
	tests := []string{"/privacy-policy", "/terms-of-service", "/security-disclosure"}

	footer := getPage(app, "/").Body.String()
	for _, path := range tests {
		if !strings.Contains(footer, `href="`+path+`"`) {
			t.Errorf("footer does not link to %s", path)
		}

		response := getPage(app, path)
		if response.Code != http.StatusOK {
			t.Errorf("GET %s: got status %d, want %d", path, response.Code, http.StatusOK)
		}
		if !strings.Contains(response.Body.String(), "Draft") {
			t.Errorf("GET %s does not identify its content as a draft", path)
		}
	}
}
