package main

import (
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"website.tabirianalytics.com/internal/database"
	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

func TestContactSubmitPersistsSubmission(t *testing.T) {
	app, db := newTestApplication(t)

	form := url.Values{
		"full_name":           {"Jane Doe"},
		"corporate_email":     {"jane@example.com"},
		"workflow_bottleneck": {"Manual invoice reconciliation"},
	}
	response := submitForm(app, "/contact", form)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); location != "/contact?submitted=true" {
		t.Fatalf("got redirect %q, want %q", location, "/contact?submitted=true")
	}

	var fullName, email, bottleneck string
	err := db.QueryRow(`
		SELECT full_name, corporate_email, workflow_bottleneck
		FROM contact_submissions
	`).Scan(&fullName, &email, &bottleneck)
	if err != nil {
		t.Fatal(err)
	}
	if fullName != "Jane Doe" || email != "jane@example.com" || bottleneck != "Manual invoice reconciliation" {
		t.Fatalf("stored unexpected submission: %q, %q, %q", fullName, email, bottleneck)
	}
}

func TestPartnerContactSubmitPersistsSubmission(t *testing.T) {
	app, db := newTestApplication(t)

	form := url.Values{
		"agency_name":         {"Example Partners"},
		"corporate_email":     {"partner@example.com"},
		"client_volume":       {"11-50"},
		"primary_vertical":    {"Healthcare"},
		"delivery_constraint": {"Limited implementation capacity"},
	}
	response := submitForm(app, "/partner-contact", form)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); location != "/partner-contact?submitted=true" {
		t.Fatalf("got redirect %q, want %q", location, "/partner-contact?submitted=true")
	}

	var agencyName, email, clientVolume, primaryVertical, deliveryConstraint string
	err := db.QueryRow(`
		SELECT agency_name, corporate_email, client_volume, primary_vertical, delivery_constraint
		FROM partner_contact_submissions
	`).Scan(&agencyName, &email, &clientVolume, &primaryVertical, &deliveryConstraint)
	if err != nil {
		t.Fatal(err)
	}
	if agencyName != "Example Partners" ||
		email != "partner@example.com" ||
		clientVolume != "11-50" ||
		primaryVertical != "Healthcare" ||
		deliveryConstraint != "Limited implementation capacity" {
		t.Fatalf(
			"stored unexpected submission: %q, %q, %q, %q, %q",
			agencyName,
			email,
			clientVolume,
			primaryVertical,
			deliveryConstraint,
		)
	}
}

func TestContactFormValidation(t *testing.T) {
	form := ContactForm{Errors: make(map[string]string)}
	form.validate()

	for _, field := range []string{"full_name", "corporate_email", "workflow_bottleneck"} {
		if form.Errors[field] == "" {
			t.Errorf("expected validation error for %s", field)
		}
	}
}

func TestPartnerContactFormValidationRejectsUnknownVolume(t *testing.T) {
	form := PartnerContactForm{
		AgencyName:         "Example Partners",
		CorporateEmail:     "partner@example.com",
		ClientVolume:       "invalid",
		PrimaryVertical:    "Healthcare",
		DeliveryConstraint: "Limited implementation capacity",
		Errors:             make(map[string]string),
	}
	form.validate()

	if form.Errors["client_volume"] == "" {
		t.Fatal("expected validation error for client_volume")
	}
}

func newTestApplication(t *testing.T) (*application, *sql.DB) {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := database.ApplySchema(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	return &application{
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		cfg:     &config{staticDir: "./ui/static"},
		queries: sqlc.New(db),
	}, db
}

func submitForm(app *application, target string, form url.Values) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)
	return response
}
