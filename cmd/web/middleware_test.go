package main

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMiddlewareAddsRequestAndSecurityHeaders(t *testing.T) {
	app, _ := newTestApplication(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "client-request-123")
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if got := response.Header().Get("X-Request-ID"); got != "client-request-123" {
		t.Fatalf("got request ID %q", got)
	}
	for _, header := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Permissions-Policy",
		"Content-Security-Policy",
	} {
		if response.Header().Get(header) == "" {
			t.Errorf("missing security header %s", header)
		}
	}
	if response.Header().Get("Server") != "" {
		t.Error("response discloses the server implementation")
	}
}

func TestHSTSOnlyInProductionBehindHTTPSLoadBalancer(t *testing.T) {
	app, _ := newTestApplication(t)
	for _, environment := range []string{"development", "production"} {
		app.cfg.environment = environment
		// The proxy connects to the app over HTTP after terminating TLS.
		response := getPage(app, "/")
		got := response.Header().Get("Strict-Transport-Security")
		if environment == "production" && got != "max-age=31536000" {
			t.Errorf("production HSTS = %q", got)
		}
		if environment == "development" && got != "" {
			t.Errorf("development must not force HTTPS: %q", got)
		}
	}
}

func TestCSPRestrictsScriptsAndAllowsExistingIntegrations(t *testing.T) {
	app, _ := newTestApplication(t)
	for _, path := range []string{"/", "/contact", "/managed-cybersecurity-services", "/static/js/analytics.js", "/missing"} {
		csp := getPage(app, path).Header().Get("Content-Security-Policy")
		if csp != contentSecurityPolicy {
			t.Errorf("%s: unexpected CSP %q", path, csp)
		}
	}
	directives := map[string]string{}
	for _, directive := range strings.Split(contentSecurityPolicy, "; ") {
		name, value, _ := strings.Cut(directive, " ")
		directives[name] = value
	}
	for _, name := range []string{"script-src", "connect-src", "frame-src"} {
		if !strings.Contains(directives[name], "https://challenges.cloudflare.com") {
			t.Errorf("%s must allow Turnstile", name)
		}
	}
	if strings.Contains(directives["script-src"], "'unsafe-inline'") ||
		strings.Contains(directives["script-src"], "'unsafe-eval'") {
		t.Error("CSP must not allow inline or evaluated scripts")
	}
	for _, test := range []struct{ name, source string }{
		{"script-src", "https://www.googletagmanager.com"},
		{"connect-src", "https://*.google-analytics.com"},
		{"frame-src", "https://www.youtube-nocookie.com"},
		{"style-src", "'unsafe-inline'"},
	} {
		if !strings.Contains(directives[test.name], test.source) {
			t.Errorf("%s must allow %s", test.name, test.source)
		}
	}
	for _, test := range []struct{ name, value string }{
		{"default-src", "'self'"},
		{"object-src", "'none'"},
		{"frame-ancestors", "'none'"},
		{"base-uri", "'self'"},
		{"form-action", "'self'"},
	} {
		if directives[test.name] != test.value {
			t.Errorf("%s = %q, want %q", test.name, directives[test.name], test.value)
		}
	}
}

func TestDocumentedBucketCSPMatchesApp(t *testing.T) {
	t.Chdir("../..")
	document, err := os.ReadFile("DEPLOYMENT.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(document), `--custom-response-header="Content-Security-Policy: `+contentSecurityPolicy+`"`) {
		t.Error("documented backend-bucket CSP differs from the app policy")
	}
}

func TestMiddlewareRecoversPanics(t *testing.T) {
	app, _ := newTestApplication(t)
	handler := app.middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("test panic")
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "test panic") {
		t.Error("response exposes panic details")
	}
}

func TestMiddlewareUsesNormalizedRouteInLogs(t *testing.T) {
	app, _ := newTestApplication(t)
	var logs bytes.Buffer
	app.logger = slog.New(slog.NewJSONHandler(&logs, nil))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodGet,
		"/blog/private-looking-slug?email=person@example.com",
		nil,
	)

	app.routes().ServeHTTP(response, request)

	output := logs.String()
	if !strings.Contains(output, `"route":"GET /blog/{id}"`) {
		t.Fatalf("log does not contain normalized route: %s", output)
	}
	for _, sensitive := range []string{"private-looking-slug", "person@example.com"} {
		if strings.Contains(output, sensitive) {
			t.Errorf("log contains request data %q", sensitive)
		}
	}
}

func TestRoutesRejectUnsupportedMethods(t *testing.T) {
	app, _ := newTestApplication(t)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodPost, "/about", nil),
	)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestStaticDirectoryListingRemainsDisabled(t *testing.T) {
	app, _ := newTestApplication(t)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/static/", nil),
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestRenderBuffersTemplateFailures(t *testing.T) {
	app, _ := newTestApplication(t)
	app.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	app.templates.pages["home"] = template.Must(
		template.New("base").Parse(`{{define "base"}}partial{{template "missing" .}}{{end}}`),
	)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "partial") {
		t.Error("response contains partially rendered template output")
	}
}

func TestLeadLogsExcludePII(t *testing.T) {
	app, _ := newTestApplication(t)
	var logs bytes.Buffer
	app.logger = slog.New(slog.NewJSONHandler(&logs, nil))
	form := validContactValues()

	response := submitForm(app, "/contact", form)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusSeeOther)
	}
	for _, pii := range []string{
		form.Get("full_name"),
		form.Get("corporate_email"),
		form.Get("workflow_bottleneck"),
	} {
		if strings.Contains(logs.String(), pii) {
			t.Errorf("logs contain submitted PII %q", pii)
		}
	}
}
