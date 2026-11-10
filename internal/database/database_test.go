package database

import (
	"context"
	"testing"
)

func TestApplySchemaMigratesExistingContactTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	_, err = db.Exec(`
		CREATE TABLE contact_submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			full_name TEXT NOT NULL,
			corporate_email TEXT NOT NULL,
			workflow_bottleneck TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplySchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	columns, err := tableColumns(context.Background(), db, "contact_submissions")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"workflow_name",
		"assessment_result",
		"assessment_score",
		"assessment_answers",
		"utm_source",
		"utm_medium",
		"utm_campaign",
		"utm_content",
		"referrer",
		"spam_status",
		"spam_reasons",
		"lead_status",
		"turnstile_verified_at",
	} {
		if !columns[name] {
			t.Errorf("missing migrated column %q", name)
		}
	}
	if !columns["created_date"] {
		t.Error("missing migrated column \"created_date\"")
	}
	if columns["created_at"] {
		t.Error("legacy column \"created_at\" should be renamed")
	}

	var spamStatus, spamReasons, leadStatus, verifiedAt string
	_, err = db.Exec(`
		INSERT INTO contact_submissions (
			full_name,
			corporate_email,
			workflow_bottleneck
		) VALUES ('Legacy User', 'legacy@example.com', 'Legacy workflow')
	`)
	if err != nil {
		t.Fatal(err)
	}
	err = db.QueryRow(`
		SELECT spam_status, spam_reasons, lead_status, turnstile_verified_at
		FROM contact_submissions
	`).Scan(&spamStatus, &spamReasons, &leadStatus, &verifiedAt)
	if err != nil {
		t.Fatal(err)
	}
	if spamStatus != "review" || spamReasons != "[]" || leadStatus != "new" || verifiedAt != "" {
		t.Fatalf(
			"unexpected migrated defaults: spam_status=%q spam_reasons=%q lead_status=%q verified_at=%q",
			spamStatus,
			spamReasons,
			leadStatus,
			verifiedAt,
		)
	}
}

func TestLeadStatusConstraints(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})
	if err := ApplySchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO contact_submissions (
			full_name,
			corporate_email,
			workflow_bottleneck,
			spam_status
		) VALUES ('Invalid User', 'invalid@example.com', 'Invalid workflow', 'unknown')
	`)
	if err == nil {
		t.Fatal("expected invalid spam_status to violate the schema constraint")
	}
}

func TestApplySchemaCreatesEmptyLeadSubmissionsTable(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := ApplySchema(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	columns, err := tableColumns(t.Context(), db, "lead_submissions")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{
		"submission_type",
		"email",
		"consent_recorded_date",
		"submitted_date",
		"next_attempt_date",
		"last_attempt_date",
		"synced_date",
		"sync_status",
		"sync_attempts",
		"hubspot_contact_id",
	} {
		if !columns[name] {
			t.Errorf("missing lead_submissions column %q", name)
		}
	}
	for _, name := range []string{
		"consent_recorded_at",
		"submitted_at",
		"next_attempt_at",
		"last_attempt_at",
		"synced_at",
	} {
		if columns[name] {
			t.Errorf("lead_submissions should not contain superseded column %q", name)
		}
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM lead_submissions").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d lead submissions after migration, want 0", count)
	}
}

func TestLeadSubmissionsQueueDefaultsAndConstraints(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	if err := ApplySchema(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO lead_submissions (submission_type, email)
		VALUES ('commercial_lead', 'lead@example.com')
	`)
	if err != nil {
		t.Fatal(err)
	}

	var syncStatus string
	var syncAttempts, consentGiven int
	var submittedDate string
	if err := db.QueryRow(`
		SELECT sync_status, sync_attempts, consent_given, submitted_date
		FROM lead_submissions
	`).Scan(&syncStatus, &syncAttempts, &consentGiven, &submittedDate); err != nil {
		t.Fatal(err)
	}
	if syncStatus != "pending" || syncAttempts != 0 || consentGiven != 0 || submittedDate == "" {
		t.Fatalf(
			"unexpected queue defaults: sync_status=%q sync_attempts=%d consent_given=%d submitted_date=%q",
			syncStatus,
			syncAttempts,
			consentGiven,
			submittedDate,
		)
	}

	if _, err := db.Exec(`
		INSERT INTO lead_submissions (submission_type, email)
		VALUES ('unknown', 'invalid@example.com')
	`); err == nil {
		t.Fatal("expected invalid submission_type to violate the schema constraint")
	}
}

func TestApplySchemaRecordsMigrationsIdempotently(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	for range 2 {
		if err := ApplySchema(t.Context(), db); err != nil {
			t.Fatal(err)
		}
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(migrations) {
		t.Fatalf("got %d migration records, want %d", count, len(migrations))
	}
}
