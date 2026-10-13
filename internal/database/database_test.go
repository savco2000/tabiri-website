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
	} {
		if !columns[name] {
			t.Errorf("missing migrated column %q", name)
		}
	}
}
