package main

import (
	"database/sql"
	"html/template"
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

func TestContactTemplateParses(t *testing.T) {
	t.Chdir("../..")
	_, err := template.ParseFiles(
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/contact.tmpl",
		"./ui/html/partials/footer.tmpl",
	)
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestAssessmentContactSubmitPersistsAttribution(t *testing.T) {
	app, db := newTestApplication(t)
	answers := `{"frequency":{"score":2,"label":"Daily or continuously"},"consistency":{"score":2,"label":"Clear rules and standard steps"},"systems":{"score":2,"label":"In connected business software"},"impact":{"score":2,"label":"Backlogs, errors, or lost capacity"},"oversight":{"score":1,"label":"An owner could be assigned"}}`
	form := url.Values{
		"full_name":           {"Jane Doe"},
		"corporate_email":     {"jane@example.com"},
		"workflow_bottleneck": {"Monthly invoice reconciliation"},
		"workflow_name":       {"Monthly invoice reconciliation"},
		"assessment_result":   {"strong"},
		"assessment_score":    {"9"},
		"assessment_answers":  {answers},
		"utm_source":          {"email_signature"},
		"utm_medium":          {"employee_email"},
		"utm_campaign":        {"automation_readiness"},
		"utm_content":         {"victor"},
		"referrer":            {"https://mail.example.com/"},
	}
	response := submitForm(app, "/contact", form)

	if response.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusSeeOther)
	}
	if location := response.Header().Get("Location"); location != "/contact?submitted=true&lead_source=automation-assessment" {
		t.Fatalf("got redirect %q", location)
	}

	var workflowName, result, score, storedAnswers, source, medium, campaign, content, referrer string
	err := db.QueryRow(`
			SELECT workflow_name, assessment_result, assessment_score, assessment_answers,
			       utm_source, utm_medium, utm_campaign, utm_content, referrer
			FROM contact_submissions
		`).Scan(&workflowName, &result, &score, &storedAnswers, &source, &medium, &campaign, &content, &referrer)
	if err != nil {
		t.Fatal(err)
	}
	if workflowName != "Monthly invoice reconciliation" || result != "strong" || score != "9" ||
		storedAnswers != answers || source != "email_signature" || medium != "employee_email" ||
		campaign != "automation_readiness" || content != "victor" ||
		referrer != "https://mail.example.com/" {
		t.Fatalf("stored unexpected assessment attribution")
	}
}

func TestAssessmentContactSubmitRejectsMalformedContext(t *testing.T) {
	app, db := newTestApplication(t)
	form := url.Values{
		"full_name":           {"Jane Doe"},
		"corporate_email":     {"jane@example.com"},
		"workflow_bottleneck": {"Invoice reconciliation"},
		"assessment_result":   {"strong"},
		"assessment_score":    {"99"},
		"assessment_answers":  {`{"frequency":{"score":2,"label":"Daily"}}`},
	}
	response := submitForm(app, "/contact", form)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM contact_submissions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("stored %d malformed submissions, want 0", count)
	}
}

func TestAssessmentContactSubmitRejectsMismatchedResult(t *testing.T) {
	app, db := newTestApplication(t)
	answers := `{"frequency":{"score":0,"label":"Monthly"},"consistency":{"score":0,"label":"Unique"},"systems":{"score":0,"label":"Offline"},"impact":{"score":0,"label":"Low"},"oversight":{"score":0,"label":"No owner"}}`
	form := url.Values{
		"full_name":           {"Jane Doe"},
		"corporate_email":     {"jane@example.com"},
		"workflow_bottleneck": {"Invoice reconciliation"},
		"assessment_result":   {"strong"},
		"assessment_score":    {"0"},
		"assessment_answers":  {answers},
	}
	response := submitForm(app, "/contact", form)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM contact_submissions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("stored %d mismatched submissions, want 0", count)
	}
}

func TestContactPrefillsAssessmentContext(t *testing.T) {
	app, _ := newTestApplication(t)
	answers := `{"frequency":{"score":2,"label":"Daily"},"consistency":{"score":2,"label":"Standard"},"systems":{"score":2,"label":"Connected"},"impact":{"score":2,"label":"High"},"oversight":{"score":1,"label":"Assigned"}}`
	query := url.Values{
		"workflow_name":      {"Invoice reconciliation"},
		"assessment_result":  {"strong"},
		"assessment_score":   {"9"},
		"assessment_answers": {answers},
		"utm_source":         {"linkedin"},
	}
	request := httptest.NewRequest(http.MethodGet, "/contact?"+query.Encode(), nil)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{
		"Discuss Your Automation Assessment",
		"Invoice reconciliation",
		`name="utm_source" type="hidden" value="linkedin"`,
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response does not contain %q", expected)
		}
	}
}

func TestContactSuccessIncludesWhatHappensNext(t *testing.T) {
	app, _ := newTestApplication(t)
	request := httptest.NewRequest(http.MethodGet, "/contact?submitted=true", nil)
	response := httptest.NewRecorder()

	app.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", response.Code, http.StatusOK)
	}
	body := response.Body.String()
	for _, expected := range []string{
		"Thank you for contacting us",
		"What Happens Next",
		"We review your workflow bottleneck and current operating constraints.",
		"We recommend the right digital associate or automation pathway.",
		"We scope a secure deployment plan tailored to your environment.",
	} {
		if !strings.Contains(body, expected) {
			t.Errorf("response does not contain %q", expected)
		}
	}
	if got := strings.Count(body, "What Happens Next"); got != 1 {
		t.Errorf("What Happens Next appears %d times, want 1", got)
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
	t.Chdir("../..")

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
