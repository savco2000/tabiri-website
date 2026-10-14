package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestContactSubmitRequiresTurnstileToken(t *testing.T) {
	app, db := newTestApplication(t)
	form := validContactValues()
	addFormProof(t, app, form, "contact")

	response := submitRawForm(app, "/contact", form)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	assertSubmissionCount(t, db, "contact_submissions", 0)
}

func TestContactSubmitRequiresSignedFormProof(t *testing.T) {
	app, db := newTestApplication(t)
	form := validContactValues()
	form.Set("cf-turnstile-response", "test-token")

	response := submitRawForm(app, "/contact", form)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	assertSubmissionCount(t, db, "contact_submissions", 0)
}

func TestContactSubmitRejectsOversizedBody(t *testing.T) {
	app, db := newTestApplication(t)
	request := httptest.NewRequest(
		http.MethodPost,
		"/contact",
		strings.NewReader("full_name="+strings.Repeat("a", maxFormBodyBytes+1)),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
	assertSubmissionCount(t, db, "contact_submissions", 0)
}

func validContactValues() url.Values {
	return url.Values{
		"full_name":           {"Jane Doe"},
		"corporate_email":     {"jane@example.com"},
		"workflow_bottleneck": {"Manual invoice reconciliation"},
	}
}

func addFormProof(t *testing.T, app *application, form url.Values, formName string) {
	t.Helper()
	proof, err := app.formProof.Create(formName)
	if err != nil {
		t.Fatal(err)
	}
	form.Set("form_proof", proof)
}

func submitRawForm(app *application, target string, form url.Values) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	app.routes().ServeHTTP(response, request)
	return response
}
