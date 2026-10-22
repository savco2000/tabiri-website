package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTurnstileHTTPVerifierAcceptsValidResponse(t *testing.T) {
	now := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.PostForm.Get("secret") != "secret-key" {
			t.Errorf("got secret %q", r.PostForm.Get("secret"))
		}
		if r.PostForm.Get("response") != "submitted-token" {
			t.Errorf("got response token %q", r.PostForm.Get("response"))
		}
		if r.PostForm.Get("idempotency_key") == "" {
			t.Error("missing idempotency key")
		}
		json.NewEncoder(w).Encode(turnstileResponse{
			Success:     true,
			ChallengeTS: now.Add(-time.Minute).Format(time.RFC3339),
			Hostname:    "www.tabirianalytics.com",
			Action:      "contact_submit",
		})
	}))
	t.Cleanup(server.Close)

	verifier := newTurnstileHTTPVerifier(
		"secret-key",
		server.URL,
		map[string]struct{}{"www.tabirianalytics.com": {}},
		false,
	)
	verifier.now = func() time.Time { return now }
	result, err := verifier.Verify(t.Context(), "submitted-token", "contact_submit")
	if err != nil {
		t.Fatal(err)
	}
	if result.Hostname != "www.tabirianalytics.com" || result.Action != "contact_submit" {
		t.Fatalf("unexpected verification result: %#v", result)
	}
}

func TestTurnstileHTTPVerifierRejectsInvalidClaims(t *testing.T) {
	now := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		response turnstileResponse
	}{
		{
			name: "unsuccessful",
			response: turnstileResponse{
				ErrorCodes: []string{"invalid-input-response"},
			},
		},
		{
			name: "wrong action",
			response: turnstileResponse{
				Success:     true,
				ChallengeTS: now.Format(time.RFC3339),
				Hostname:    "www.tabirianalytics.com",
				Action:      "partner_contact_submit",
			},
		},
		{
			name: "wrong hostname",
			response: turnstileResponse{
				Success:     true,
				ChallengeTS: now.Format(time.RFC3339),
				Hostname:    "attacker.example",
				Action:      "contact_submit",
			},
		},
		{
			name: "expired challenge",
			response: turnstileResponse{
				Success:     true,
				ChallengeTS: now.Add(-11 * time.Minute).Format(time.RFC3339),
				Hostname:    "www.tabirianalytics.com",
				Action:      "contact_submit",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				json.NewEncoder(w).Encode(tt.response)
			}))
			t.Cleanup(server.Close)
			verifier := newTurnstileHTTPVerifier(
				"secret-key",
				server.URL,
				map[string]struct{}{"www.tabirianalytics.com": {}},
				false,
			)
			verifier.now = func() time.Time { return now }
			_, err := verifier.Verify(t.Context(), "submitted-token", "contact_submit")
			if !errors.Is(err, errTurnstileRejected) {
				t.Fatalf("got %v, want errTurnstileRejected", err)
			}
		})
	}
}

func TestTurnstileHTTPVerifierAcceptsDummyTestClaimsInTestMode(t *testing.T) {
	now := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := turnstileResponse{
			Success:     true,
			ChallengeTS: now.Format(time.RFC3339),
			Hostname:    "example.com",
		}
		response.Metadata.ResultWithTestingKey = true
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(server.Close)

	verifier := newTurnstileHTTPVerifier(
		testTurnstileSecretKey,
		server.URL,
		map[string]struct{}{"localhost": {}},
		true,
	)
	verifier.now = func() time.Time { return now }
	if _, err := verifier.Verify(t.Context(), "XXXX.DUMMY.TOKEN.XXXX", "contact_submit"); err != nil {
		t.Fatal(err)
	}
}

func TestTurnstileHTTPVerifierRequiresTestingMarkerInTestMode(t *testing.T) {
	now := time.Date(2026, time.October, 14, 16, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(turnstileResponse{
			Success:     true,
			ChallengeTS: now.Format(time.RFC3339),
			Hostname:    "example.com",
		})
	}))
	t.Cleanup(server.Close)

	verifier := newTurnstileHTTPVerifier(
		testTurnstileSecretKey,
		server.URL,
		map[string]struct{}{"localhost": {}},
		true,
	)
	verifier.now = func() time.Time { return now }
	_, err := verifier.Verify(t.Context(), "submitted-token", "contact_submit")
	if !errors.Is(err, errTurnstileRejected) {
		t.Fatalf("got %v, want errTurnstileRejected", err)
	}
}

func TestTurnstileHTTPVerifierSurfacesProviderFailures(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name: "non-200 response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "unavailable", http.StatusServiceUnavailable)
			},
		},
		{
			name: "malformed response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte("{"))
			},
		},
		{
			name: "oversized response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte(strings.Repeat("x", maxTurnstileResponseBytes+1)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			t.Cleanup(server.Close)
			verifier := newTurnstileHTTPVerifier(
				"secret-key",
				server.URL,
				map[string]struct{}{"www.tabirianalytics.com": {}},
				false,
			)
			_, err := verifier.Verify(t.Context(), "submitted-token", "contact_submit")
			if !errors.Is(err, errTurnstileUnavailable) {
				t.Fatalf("got %v, want errTurnstileUnavailable", err)
			}
		})
	}
}

func TestTurnstileHTTPVerifierPreservesCancellation(t *testing.T) {
	verifier := newTurnstileHTTPVerifier(
		"secret-key",
		"https://example.invalid",
		map[string]struct{}{"www.tabirianalytics.com": {}},
		false,
	)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := verifier.Verify(ctx, "submitted-token", "contact_submit")

	if !errors.Is(err, errTurnstileUnavailable) {
		t.Fatalf("got %v, want errTurnstileUnavailable", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context cancellation preserved", err)
	}
}
