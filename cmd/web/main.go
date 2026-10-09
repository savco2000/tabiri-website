package main

import (
	"context"
	"flag" // New import
	"log/slog"
	"net/http"
	"os"

	"website.tabirianalytics.com/internal/database"
	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include the structured logger, but we'll
// add more to this as the build progresses.
type application struct {
	logger  *slog.Logger
	cfg     *config
	queries *sqlc.Queries
}

type config struct {
	addr      string
	staticDir string
	dbPath    string
}

func main() {

	var cfg config

	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	flag.StringVar(&cfg.dbPath, "db-path", "./tabiri.db", "Path to SQLite database")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	db, err := database.Open(cfg.dbPath)
	if err != nil {
		logger.Error("opening database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.ApplySchema(context.Background(), db); err != nil {
		logger.Error("applying database schema", "error", err)
		os.Exit(1)
	}

	app := &application{
		logger:  logger,
		cfg:     &cfg,
		queries: sqlc.New(db),
	}

	logger.Info("starting server", "addr", cfg.addr)
	// Call the new app.routes() method to get the servemux containing our routes,
	// and pass that to http.ListenAndServe().
	err = http.ListenAndServe(cfg.addr, app.routes())

	logger.Error(err.Error())
	os.Exit(1)
}
