package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

const (
	hubspotMaxAttempts = 8
	hubspotMaxBackoff  = time.Hour
	hubspotSyncPacing  = 250 * time.Millisecond
)

type leadSynchronizer interface {
	Sync(context.Context, sqlc.LeadSubmission) (string, error)
}

type hubspotWorker struct {
	queries        *sqlc.Queries
	synchronizer   leadSynchronizer
	logger         *slog.Logger
	telemetry      *telemetry
	pollInterval   time.Duration
	requestTimeout time.Duration
	now            func() time.Time
}

func (w *hubspotWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if err := w.drain(ctx); err != nil && !errors.Is(err, context.Canceled) {
			w.logger.ErrorContext(ctx, "run HubSpot synchronization", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (w *hubspotWorker) drain(ctx context.Context) error {
	for {
		processed, err := w.processNext(ctx)
		if err != nil || !processed {
			return err
		}
		timer := time.NewTimer(hubspotSyncPacing)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (w *hubspotWorker) processNext(ctx context.Context) (bool, error) {
	submission, err := w.queries.ClaimLeadSubmission(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim lead submission: %w", err)
	}

	started := w.now()
	requestCtx, cancel := context.WithTimeout(ctx, w.requestTimeout)
	contactID, syncErr := w.synchronizer.Sync(requestCtx, submission)
	cancel()
	if syncErr == nil {
		_, err = w.queries.MarkLeadSubmissionSynced(ctx, sqlc.MarkLeadSubmissionSyncedParams{
			HubspotContactID: contactID,
			ID:               submission.ID,
		})
		if err != nil {
			return true, fmt.Errorf("mark lead submission %d synced: %w", submission.ID, err)
		}
		w.telemetry.recordHubSpotSync(
			ctx,
			submission.SubmissionType,
			"synced",
			w.now().Sub(started),
		)
		return true, nil
	}

	retryAt, retryable := w.retrySchedule(submission.SyncAttempts, syncErr)
	nextAttempt := sql.NullString{}
	outcome := "permanent_failure"
	if retryable {
		nextAttempt = sql.NullString{
			String: retryAt.UTC().Format("2006-01-02 15:04:05"),
			Valid:  true,
		}
		outcome = "retry_scheduled"
	}
	_, err = w.queries.MarkLeadSubmissionFailed(ctx, sqlc.MarkLeadSubmissionFailedParams{
		NextAttemptDate: nextAttempt,
		LastSyncError:   syncErr.Error(),
		ID:              submission.ID,
	})
	if err != nil {
		return true, fmt.Errorf(
			"mark lead submission %d failed after synchronization error: %w",
			submission.ID,
			err,
		)
	}
	w.telemetry.recordHubSpotSync(
		ctx,
		submission.SubmissionType,
		outcome,
		w.now().Sub(started),
	)
	w.logger.WarnContext(
		ctx,
		"HubSpot synchronization failed",
		"submission_id", submission.ID,
		"submission_type", submission.SubmissionType,
		"attempt", submission.SyncAttempts,
		"retry_scheduled", retryable,
		"error", syncErr,
	)
	return true, nil
}

func (w *hubspotWorker) retrySchedule(attempt int64, err error) (time.Time, bool) {
	if attempt >= hubspotMaxAttempts || !retryableHubSpotError(err) {
		return time.Time{}, false
	}
	delay := time.Minute * time.Duration(1<<uint(max(attempt-1, 0)))
	if delay > hubspotMaxBackoff {
		delay = hubspotMaxBackoff
	}
	var apiErr *hubspotAPIError
	if errors.As(err, &apiErr) && apiErr.retryAfter > delay {
		delay = apiErr.retryAfter
	}
	return w.now().Add(delay), true
}

func retryableHubSpotError(err error) bool {
	var apiErr *hubspotAPIError
	if !errors.As(err, &apiErr) {
		return true
	}
	return apiErr.statusCode == http.StatusRequestTimeout ||
		apiErr.statusCode == http.StatusTooEarly ||
		apiErr.statusCode == http.StatusConflict ||
		apiErr.statusCode == http.StatusTooManyRequests ||
		apiErr.statusCode >= http.StatusInternalServerError
}
