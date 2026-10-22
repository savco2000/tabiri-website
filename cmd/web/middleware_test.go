package main

import (
	"bytes"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	} {
		if response.Header().Get(header) == "" {
			t.Errorf("missing security header %s", header)
		}
	}
	if response.Header().Get("Server") != "" {
		t.Error("response discloses the server implementation")
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
