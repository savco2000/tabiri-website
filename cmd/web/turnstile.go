package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTurnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	maxTurnstileResponseBytes = 64 << 10
	testTurnstileSiteKey      = "1x00000000000000000000BB"
	testTurnstileVisibleKey   = "1x00000000000000000000AA"
	testTurnstileSecretKey    = "1x0000000000000000000000000000000AA"
)

var (
	errTurnstileRejected    = errors.New("turnstile rejected submission")
	errTurnstileUnavailable = errors.New("turnstile verification unavailable")
)

type turnstileVerification struct {
	Hostname    string
	Action      string
	ChallengeAt time.Time
}

type turnstileVerifier interface {
	Verify(ctx context.Context, token, expectedAction string) (turnstileVerification, error)
}

type turnstileHTTPVerifier struct {
	client           *http.Client
	secretKey        string
	verifyURL        string
	allowedHostnames map[string]struct{}
	now              func() time.Time
	testMode         bool
	telemetry        *telemetry
}

type turnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Action      string   `json:"action"`
	ErrorCodes  []string `json:"error-codes"`
	Metadata    struct {
		ResultWithTestingKey bool `json:"result_with_testing_key"`
	} `json:"metadata"`
}

func newTurnstileHTTPVerifier(
	secretKey string,
	verifyURL string,
	allowedHostnames map[string]struct{},
	testMode bool,
) *turnstileHTTPVerifier {
	if verifyURL == "" {
		verifyURL = defaultTurnstileVerifyURL
	}
	return &turnstileHTTPVerifier{
		client:           &http.Client{Timeout: 5 * time.Second},
		secretKey:        secretKey,
		verifyURL:        verifyURL,
		allowedHostnames: allowedHostnames,
		now:              time.Now,
		testMode:         testMode,
	}
}

func (v *turnstileHTTPVerifier) Verify(
	ctx context.Context,
	token,
	expectedAction string,
) (verification turnstileVerification, resultErr error) {
	tracer := otel.Tracer(instrumentationName)
	if v.telemetry != nil {
		tracer = v.telemetry.tracer
	}
	ctx, span := tracer.Start(
		ctx,
		"turnstile verify",
		trace.WithAttributes(attribute.String("turnstile.expected_action", expectedAction)),
	)
	started := time.Now()
	outcome := "unavailable"
	defer func() {
		span.SetAttributes(attribute.String("turnstile.outcome", outcome))
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, outcome)
		}
		span.End()
		if v.telemetry != nil {
			v.telemetry.recordTurnstile(ctx, outcome, time.Since(started))
		}
	}()

	idempotencyKey, err := randomUUID()
	if err != nil {
		return turnstileVerification{}, fmt.Errorf("%w: generating idempotency key: %w", errTurnstileUnavailable, err)
	}

	values := url.Values{
		"secret":          {v.secretKey},
		"response":        {token},
		"idempotency_key": {idempotencyKey},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.verifyURL, strings.NewReader(values.Encode()))
	if err != nil {
		return turnstileVerification{}, fmt.Errorf("%w: creating request: %w", errTurnstileUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return turnstileVerification{}, fmt.Errorf("%w: %w", errTurnstileUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return turnstileVerification{}, fmt.Errorf("%w: unexpected status %d", errTurnstileUnavailable, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTurnstileResponseBytes+1))
	if err != nil {
		return turnstileVerification{}, fmt.Errorf("%w: reading response: %w", errTurnstileUnavailable, err)
	}
	if len(body) > maxTurnstileResponseBytes {
		return turnstileVerification{}, fmt.Errorf("%w: response exceeds size limit", errTurnstileUnavailable)
	}

	var result turnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return turnstileVerification{}, fmt.Errorf("%w: decoding response: %w", errTurnstileUnavailable, err)
	}
	if !result.Success {
		outcome = "rejected"
		return turnstileVerification{}, fmt.Errorf("%w: error codes %v", errTurnstileRejected, result.ErrorCodes)
	}
	if v.testMode {
		if !result.Metadata.ResultWithTestingKey {
			outcome = "rejected"
			return turnstileVerification{}, fmt.Errorf("%w: missing testing-key marker", errTurnstileRejected)
		}
	} else {
		if result.Action != expectedAction {
			outcome = "rejected"
			return turnstileVerification{}, fmt.Errorf("%w: action mismatch", errTurnstileRejected)
		}
		if _, ok := v.allowedHostnames[normalizeHostname(result.Hostname)]; !ok {
			outcome = "rejected"
			return turnstileVerification{}, fmt.Errorf("%w: hostname mismatch", errTurnstileRejected)
		}
	}

	challengeAt, err := time.Parse(time.RFC3339, result.ChallengeTS)
	if err != nil {
		outcome = "rejected"
		return turnstileVerification{}, fmt.Errorf("%w: invalid challenge timestamp", errTurnstileRejected)
	}
	now := v.now()
	if challengeAt.After(now.Add(30*time.Second)) || now.Sub(challengeAt) > 10*time.Minute {
		outcome = "rejected"
		return turnstileVerification{}, fmt.Errorf("%w: challenge timestamp outside allowed window", errTurnstileRejected)
	}

	outcome = "accepted"
	return turnstileVerification{
		Hostname:    result.Hostname,
		Action:      result.Action,
		ChallengeAt: challengeAt,
	}, nil
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		value[0:4],
		value[4:6],
		value[6:8],
		value[8:10],
		value[10:16],
	), nil
}

func normalizeHostname(hostname string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(hostname)), ".")
}
