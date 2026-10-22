package main

import (
	"log/slog"
	"os"
	"time"

	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

type application struct {
	logger    *slog.Logger
	cfg       *config
	queries   *sqlc.Queries
	turnstile turnstileVerifier
	formProof *formProofSigner
	templates *templateCache
	telemetry *telemetry
	now       func() time.Time
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
