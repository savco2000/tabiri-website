package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
)

//go:embed schema.sql
var schema string

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func ApplySchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return err
	}

	for _, table := range []string{"contact_submissions", "partner_contact_submissions"} {
		if err := renameCreatedColumn(ctx, db, table); err != nil {
			return err
		}
	}

	columns := map[string]string{
		"workflow_name":      "TEXT NOT NULL DEFAULT ''",
		"assessment_result":  "TEXT NOT NULL DEFAULT ''",
		"assessment_score":   "TEXT NOT NULL DEFAULT ''",
		"assessment_answers": "TEXT NOT NULL DEFAULT ''",
		"utm_source":         "TEXT NOT NULL DEFAULT ''",
		"utm_medium":         "TEXT NOT NULL DEFAULT ''",
		"utm_campaign":       "TEXT NOT NULL DEFAULT ''",
		"utm_content":        "TEXT NOT NULL DEFAULT ''",
		"referrer":           "TEXT NOT NULL DEFAULT ''",
	}

	existing, err := tableColumns(ctx, db, "contact_submissions")
	if err != nil {
		return err
	}
	for name, definition := range columns {
		if existing[name] {
			continue
		}
		if _, err := db.ExecContext(ctx, fmt.Sprintf(
			"ALTER TABLE contact_submissions ADD COLUMN %s %s",
			name,
			definition,
		)); err != nil {
			return fmt.Errorf("adding contact_submissions.%s: %w", name, err)
		}
	}

	return nil
}

func renameCreatedColumn(ctx context.Context, db *sql.DB, table string) error {
	columns, err := tableColumns(ctx, db, table)
	if err != nil {
		return err
	}
	if columns["created_date"] || !columns["created_at"] {
		return nil
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE "+table+" RENAME COLUMN created_at TO created_date"); err != nil {
		return fmt.Errorf("renaming %s.created_at: %w", table, err)
	}
	return nil
}

func tableColumns(ctx context.Context, db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, err
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}
