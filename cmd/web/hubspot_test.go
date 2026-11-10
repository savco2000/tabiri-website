package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"website.tabirianalytics.com/internal/database"
	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

func TestHubSpotSyncUpdatesContactAndNewsletterPreference(t *testing.T) {
	var mu sync.Mutex
	var requests []recordedHubSpotRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		mu.Lock()
		requests = append(requests, recordedHubSpotRequest{
			method: r.Method,
			path:   r.URL.Path,
			body:   body,
		})
		mu.Unlock()

		switch r.URL.Path {
		case "/crm/v3/objects/contacts/search":
			writeJSON(t, w, map[string]any{
				"results": []any{
					map[string]any{
						"id": "123",
						"properties": map[string]string{
							hubspotEngagementProperty: "commercial_lead",
						},
					},
				},
			})
		case "/crm/v3/objects/contacts/123":
			writeJSON(t, w, map[string]string{"id": "123"})
		case "/communication-preferences/v3/subscribe":
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestHubSpotClient(server.URL)
	contactID, err := client.Sync(t.Context(), sqlc.LeadSubmission{
		SubmissionType: "newsletter",
		Email:          " Person@Example.com ",
		FirstName:      "Ada",
		ConsentGiven:   1,
		SubmittedDate:  "2026-03-20 12:34:56",
	})
	if err != nil {
		t.Fatal(err)
	}
	if contactID != "123" {
		t.Fatalf("got contact ID %q, want 123", contactID)
	}
	if len(requests) != 3 {
		t.Fatalf("got %d requests, want 3", len(requests))
	}
	if requests[1].method != http.MethodPatch {
		t.Fatalf("got contact method %s, want PATCH", requests[1].method)
	}
	properties := requests[1].body["properties"].(map[string]any)
	if properties[hubspotEngagementProperty] != "commercial_lead;newsletter" {
		t.Fatalf("got engagement values %q", properties[hubspotEngagementProperty])
	}
	if properties["email"] != "person@example.com" {
		t.Fatalf("got normalized email %q", properties["email"])
	}
	if requests[2].body["legalBasis"] != "CONSENT_WITH_NOTICE" {
		t.Fatalf("got legal basis %q", requests[2].body["legalBasis"])
	}
	if requests[2].body["subscriptionId"] != "4078871385" {
		t.Fatalf("got subscription ID %q", requests[2].body["subscriptionId"])
	}
}

func TestHubSpotSyncCreatesCommercialContactWithoutMarketingPreference(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/crm/v3/objects/contacts/search":
			writeJSON(t, w, map[string]any{"results": []any{}})
		case "/crm/v3/objects/contacts":
			var request hubspotContactRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.Properties["firstname"] != "Ada" ||
				request.Properties["lastname"] != "Lovelace" {
				t.Fatalf("unexpected name properties: %#v", request.Properties)
			}
			if request.Properties[hubspotEngagementProperty] != "commercial_lead" {
				t.Fatalf("unexpected engagement property: %#v", request.Properties)
			}
			writeJSON(t, w, map[string]string{"id": "456"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestHubSpotClient(server.URL)
	contactID, err := client.Sync(t.Context(), sqlc.LeadSubmission{
		SubmissionType:     "commercial_lead",
		Email:              "ada@example.com",
		FullName:           "Ada Lovelace",
		WorkflowBottleneck: "Manual reporting",
		SubmittedDate:      "2026-03-20T12:34:56Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if contactID != "456" {
		t.Fatalf("got contact ID %q, want 456", contactID)
	}
	if len(paths) != 2 {
		t.Fatalf("got request paths %v, want search and create only", paths)
	}
}

func TestHubSpotSyncDoesNotSubscribeNewsletterWithoutConsent(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/crm/v3/objects/contacts/search":
			writeJSON(t, w, map[string]any{"results": []any{}})
		case "/crm/v3/objects/contacts":
			writeJSON(t, w, map[string]string{"id": "456"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := newTestHubSpotClient(server.URL)
	if _, err := client.Sync(t.Context(), sqlc.LeadSubmission{
		SubmissionType: "newsletter",
		Email:          "news@example.com",
		ConsentGiven:   0,
		SubmittedDate:  "2026-03-20T12:34:56Z",
	}); err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("got request paths %v, want search and create only", paths)
	}
}

func TestHubSpotErrorClassification(t *testing.T) {
	for _, status := range []int{
		http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusConflict,
		http.StatusTooManyRequests,
		http.StatusBadGateway,
	} {
		if !retryableHubSpotError(&hubspotAPIError{statusCode: status}) {
			t.Errorf("expected status %d to be retryable", status)
		}
	}
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden} {
		if retryableHubSpotError(&hubspotAPIError{statusCode: status}) {
			t.Errorf("expected status %d to be permanent", status)
		}
	}
	if !retryableHubSpotError(errors.New("network failure")) {
		t.Fatal("expected transport errors to be retryable")
	}
}

func TestHubSpotWorkerMarksSubmissionSynced(t *testing.T) {
	worker, db, synchronizer := newTestHubSpotWorker(t)
	submissionID := insertPendingSubmission(t, db, "commercial_lead", "lead@example.com")
	synchronizer.contactID = "789"

	processed, err := worker.processNext(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !processed {
		t.Fatal("expected a submission to be processed")
	}

	var status, contactID string
	var attempts int
	if err := db.QueryRow(
		`SELECT sync_status, sync_attempts, hubspot_contact_id
		 FROM lead_submissions WHERE id = ?`,
		submissionID,
	).Scan(&status, &attempts, &contactID); err != nil {
		t.Fatal(err)
	}
	if status != "synced" || attempts != 1 || contactID != "789" {
		t.Fatalf("got status=%q attempts=%d contact=%q", status, attempts, contactID)
	}
}

func TestHubSpotWorkerSchedulesRetryAndStopsAtPermanentFailure(t *testing.T) {
	worker, db, synchronizer := newTestHubSpotWorker(t)
	retryID := insertPendingSubmission(t, db, "partner", "partner@example.com")
	synchronizer.err = &hubspotAPIError{statusCode: http.StatusTooManyRequests}

	if _, err := worker.processNext(t.Context()); err != nil {
		t.Fatal(err)
	}
	var retryStatus string
	var nextAttempt sql.NullString
	if err := db.QueryRow(
		`SELECT sync_status, next_attempt_date FROM lead_submissions WHERE id = ?`,
		retryID,
	).Scan(&retryStatus, &nextAttempt); err != nil {
		t.Fatal(err)
	}
	if retryStatus != "failed" || !nextAttempt.Valid {
		t.Fatalf("got status=%q next_attempt=%#v", retryStatus, nextAttempt)
	}

	permanentID := insertPendingSubmission(t, db, "newsletter", "news@example.com")
	synchronizer.err = &hubspotAPIError{statusCode: http.StatusBadRequest}
	if _, err := worker.processNext(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(
		`SELECT sync_status, next_attempt_date FROM lead_submissions WHERE id = ?`,
		permanentID,
	).Scan(&retryStatus, &nextAttempt); err != nil {
		t.Fatal(err)
	}
	if retryStatus != "failed" || nextAttempt.Valid {
		t.Fatalf("got status=%q next_attempt=%#v", retryStatus, nextAttempt)
	}
	if _, err := worker.queries.ClaimLeadSubmission(t.Context()); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("got claim error %v, want sql.ErrNoRows", err)
	}
}

func TestClaimLeadSubmissionRecoversStaleProcessingOnly(t *testing.T) {
	_, db, _ := newTestHubSpotWorker(t)
	queries := sqlc.New(db)
	staleID := insertPendingSubmission(t, db, "commercial_lead", "stale@example.com")
	freshID := insertPendingSubmission(t, db, "commercial_lead", "fresh@example.com")
	if _, err := db.Exec(
		`UPDATE lead_submissions
		 SET sync_status = 'processing', last_attempt_date = '2000-01-01 00:00:00'
		 WHERE id = ?`,
		staleID,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`UPDATE lead_submissions
		 SET sync_status = 'processing', last_attempt_date = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		freshID,
	); err != nil {
		t.Fatal(err)
	}

	claimed, err := queries.ClaimLeadSubmission(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if claimed.ID != staleID || claimed.SyncAttempts != 1 {
		t.Fatalf("claimed ID=%d attempts=%d, want stale ID=%d attempts=1", claimed.ID, claimed.SyncAttempts, staleID)
	}
	if _, err := queries.ClaimLeadSubmission(t.Context()); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("got second claim error %v, want sql.ErrNoRows", err)
	}
}

func newTestHubSpotClient(baseURL string) *hubspotClient {
	return &hubspotClient{
		httpClient:               &http.Client{Timeout: time.Second},
		baseURL:                  baseURL,
		serviceKey:               "test-service-key",
		newsletterSubscriptionID: "4078871385",
	}
}

type recordedHubSpotRequest struct {
	method string
	path   string
	body   map[string]any
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

type fakeLeadSynchronizer struct {
	contactID string
	err       error
}

func (s *fakeLeadSynchronizer) Sync(
	_ context.Context,
	_ sqlc.LeadSubmission,
) (string, error) {
	return s.contactID, s.err
}

func newTestHubSpotWorker(
	t *testing.T,
) (*hubspotWorker, *sql.DB, *fakeLeadSynchronizer) {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.ApplySchema(t.Context(), db); err != nil {
		t.Fatal(err)
	}
	telemetry, err := newTelemetry(t.Context(), &config{environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { telemetry.shutdown(context.Background()) })

	synchronizer := &fakeLeadSynchronizer{}
	return &hubspotWorker{
		queries:        sqlc.New(db),
		synchronizer:   synchronizer,
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		telemetry:      telemetry,
		pollInterval:   time.Millisecond,
		requestTimeout: time.Second,
		now: func() time.Time {
			return time.Date(2099, 3, 20, 12, 0, 0, 0, time.UTC)
		},
	}, db, synchronizer
}

func insertPendingSubmission(
	t *testing.T,
	db *sql.DB,
	submissionType string,
	email string,
) int64 {
	t.Helper()
	result, err := db.Exec(
		`INSERT INTO lead_submissions (submission_type, email) VALUES (?, ?)`,
		submissionType,
		email,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestLoadHubSpotConfig(t *testing.T) {
	t.Setenv("HUBSPOT_SERVICE_KEY", "")
	cfg := config{environment: "production"}
	if err := loadHubSpotConfig(&cfg); err == nil {
		t.Fatal("expected production configuration without a service key to fail")
	}

	t.Setenv("HUBSPOT_SERVICE_KEY", "secret")
	t.Setenv("HUBSPOT_API_BASE_URL", "https://example.test/")
	if err := loadHubSpotConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.hubspotEnabled || strings.HasSuffix(cfg.hubspotAPIBaseURL, "/") {
		t.Fatalf("unexpected HubSpot configuration: %#v", cfg)
	}
}

func TestHubSpotWorkerStopsOnCancellation(t *testing.T) {
	worker, _, _ := newTestHubSpotWorker(t)
	worker.pollInterval = time.Hour
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.run(ctx)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}
}
