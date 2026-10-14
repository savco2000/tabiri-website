package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type spamStatus string

const (
	spamStatusAccepted spamStatus = "accepted"
	spamStatusReview   spamStatus = "review"
	spamStatusRejected spamStatus = "rejected"
)

type leadStatus string

const (
	leadStatusNew          leadStatus = "new"
	leadStatusContacted    leadStatus = "contacted"
	leadStatusQualified    leadStatus = "qualified"
	leadStatusDisqualified leadStatus = "disqualified"
	leadStatusArchived     leadStatus = "archived"
)

type leadVerification struct {
	Turnstile  turnstileVerification
	VerifiedAt time.Time
}

type leadVerificationContextKey struct{}

type leadProtectionErrorRenderer func(http.ResponseWriter, *http.Request, int, string)

func (app *application) protectLeadSubmission(
	formName string,
	action string,
	renderError leadProtectionErrorRenderer,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)
		if err := r.ParseForm(); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				app.clientError(w, http.StatusRequestEntityTooLarge)
				return
			}
			app.clientError(w, http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(r.PostForm.Get("website")) != "" {
			app.logger.Warn("lead submission rejected", "form", formName, "reason", "honeypot")
			renderError(w, r, http.StatusUnprocessableEntity, verificationFailureMessage)
			return
		}
		if err := app.formProof.Verify(r.PostForm.Get("form_proof"), formName); err != nil {
			app.logger.Warn("lead submission rejected", "form", formName, "reason", "form_proof")
			renderError(w, r, http.StatusUnprocessableEntity, verificationFailureMessage)
			return
		}

		token := strings.TrimSpace(r.PostForm.Get("cf-turnstile-response"))
		if token == "" {
			renderError(w, r, http.StatusUnprocessableEntity, verificationFailureMessage)
			return
		}
		verification, err := app.turnstile.Verify(r.Context(), token, action)
		if err != nil {
			if errors.Is(err, errTurnstileRejected) {
				app.logger.Warn("lead submission rejected", "form", formName, "reason", "turnstile")
				renderError(w, r, http.StatusUnprocessableEntity, verificationFailureMessage)
				return
			}
			app.logger.Error("turnstile verification unavailable", "form", formName, "error", err)
			renderError(w, r, http.StatusServiceUnavailable, verificationUnavailableMessage)
			return
		}

		ctx := context.WithValue(r.Context(), leadVerificationContextKey{}, leadVerification{
			Turnstile:  verification,
			VerifiedAt: time.Now().UTC(),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requireLeadVerification(ctx context.Context) (leadVerification, error) {
	verification, ok := ctx.Value(leadVerificationContextKey{}).(leadVerification)
	if !ok {
		return leadVerification{}, fmt.Errorf("missing lead verification context")
	}
	return verification, nil
}

const (
	verificationFailureMessage     = "We could not verify your submission. Please refresh the page and try again."
	verificationUnavailableMessage = "Submission verification is temporarily unavailable. Please try again shortly."
)
