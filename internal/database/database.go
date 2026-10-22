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
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	// SQLite serializes writes. A single connection also keeps PRAGMA settings
	// consistent and makes in-memory databases behave predictably in tests.
	db.SetMaxOpenConns(1)

	for _, pragma := range []string{
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("configure sqlite database: %w", err)
		}
	}

	return db, nil
}

type migration struct {
	version int
	apply   func(context.Context, *sql.Tx) error
}

var migrations = []migration{
	{version: 1, apply: createCurrentSchema},
	{version: 2, apply: migrateLegacyLeadSchema},
}

func ApplySchema(ctx context.Context, db *sql.DB) error {
	return ApplySchemaWithObserver(ctx, db, nil)
}

func ApplySchemaWithObserver(
	ctx context.Context,
	db *sql.DB,
	onApplied func(version int),
) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	for _, migration := range migrations {
		var applied bool
		if err := db.QueryRowContext(
			ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)",
			migration.version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %d: %w", migration.version, err)
		}
		if applied {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", migration.version, err)
		}
		if err := migration.apply(ctx, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO schema_migrations (version) VALUES (?)",
			migration.version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", migration.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.version, err)
		}
		if onApplied != nil {
			onApplied(migration.version)
		}
	}

	return nil
}

func createCurrentSchema(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("create current schema: %w", err)
	}
	return nil
}

func migrateLegacyLeadSchema(ctx context.Context, tx *sql.Tx) error {
	for _, table := range []string{"contact_submissions", "partner_contact_submissions"} {
		if err := renameCreatedColumn(ctx, tx, table); err != nil {
			return err
		}
	}

	contactColumns := []columnMigration{
		{name: "workflow_name", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "assessment_result", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "assessment_score", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "assessment_answers", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "utm_source", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "utm_medium", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "utm_campaign", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "utm_content", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "referrer", definition: "TEXT NOT NULL DEFAULT ''"},
	}
	if err := addMissingColumns(ctx, tx, "contact_submissions", contactColumns); err != nil {
		return err
	}

	leadColumns := []columnMigration{
		{name: "spam_status", definition: "TEXT NOT NULL DEFAULT 'review' CHECK (spam_status IN ('accepted', 'review', 'rejected'))"},
		{name: "spam_reasons", definition: "TEXT NOT NULL DEFAULT '[]'"},
		{name: "lead_status", definition: "TEXT NOT NULL DEFAULT 'new' CHECK (lead_status IN ('new', 'contacted', 'qualified', 'disqualified', 'archived'))"},
		{name: "turnstile_verified_at", definition: "TEXT NOT NULL DEFAULT ''"},
	}
	for _, table := range []string{"contact_submissions", "partner_contact_submissions"} {
		if err := addMissingColumns(ctx, tx, table, leadColumns); err != nil {
			return err
		}
	}
	return nil
}

type columnMigration struct {
	name       string
	definition string
}

func addMissingColumns(ctx context.Context, tx *sql.Tx, table string, columns []columnMigration) error {
	existing, err := tableColumns(ctx, tx, table)
	if err != nil {
		return err
	}
	for _, column := range columns {
		if existing[column.name] {
			continue
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN %s %s",
			table,
			column.name,
			column.definition,
		)); err != nil {
			return fmt.Errorf("add %s.%s: %w", table, column.name, err)
		}
	}
	return nil
}

type databaseQuerier interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func renameCreatedColumn(ctx context.Context, db databaseQuerier, table string) error {
	columns, err := tableColumns(ctx, db, table)
	if err != nil {
		return err
	}
	if columns["created_date"] || !columns["created_at"] {
		return nil
	}
	if _, err := db.ExecContext(ctx, "ALTER TABLE "+table+" RENAME COLUMN created_at TO created_date"); err != nil {
		return fmt.Errorf("rename %s.created_at: %w", table, err)
	}
	return nil
}

func tableColumns(ctx context.Context, db databaseQuerier, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		return nil, fmt.Errorf("inspect %s columns: %w", table, err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, fmt.Errorf("scan %s column: %w", table, err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s columns: %w", table, err)
	}
	return columns, nil
}
