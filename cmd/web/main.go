package main

import (
	"flag" // New import
	"log/slog"
	"net/http"
	"os"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include the structured logger, but we'll
// add more to this as the build progresses.
type application struct {
	logger *slog.Logger
	cfg    *config
}

type config struct {
	addr      string
	staticDir string
}

func main() {

	var cfg config

	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	app := &application{
		logger: logger,
		cfg:    &cfg,
	}

	logger.Info("starting server", "addr", cfg.addr)
	// Call the new app.routes() method to get the servemux containing our routes,
	// and pass that to http.ListenAndServe().
	err := http.ListenAndServe(cfg.addr, app.routes())

	logger.Error(err.Error())
	os.Exit(1)
}
