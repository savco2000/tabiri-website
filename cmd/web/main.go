package main

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"website.tabirianalytics.com/internal/database"
	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

// Define an application struct to hold the application-wide dependencies for the
// web application. For now we'll only include the structured logger, but we'll
// add more to this as the build progresses.
type application struct {
	logger    *slog.Logger
	cfg       *config
	queries   *sqlc.Queries
	turnstile turnstileVerifier
	formProof *formProofSigner
}

type config struct {
	addr                  string
	staticDir             string
	dbPath                string
	environment           string
	turnstileSiteKey      string
	turnstileSecretKey    string
	turnstileVerifyURL    string
	turnstileAllowedHosts map[string]struct{}
	turnstileTestMode     bool
	formHMACKey           []byte
}

func main() {

	var cfg config
	environment := strings.TrimSpace(os.Getenv("APP_ENV"))
	if environment == "" {
		environment = "development"
	}

	flag.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	flag.StringVar(&cfg.dbPath, "db-path", "./tabiri.db", "Path to SQLite database")
	flag.StringVar(&cfg.environment, "environment", environment, "Application environment")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	if err := loadLeadProtectionConfig(&cfg); err != nil {
		logger.Error("loading lead protection configuration", "error", err)
		os.Exit(1)
	}

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
		turnstile: newTurnstileHTTPVerifier(
			cfg.turnstileSecretKey,
			cfg.turnstileVerifyURL,
			cfg.turnstileAllowedHosts,
			cfg.turnstileTestMode,
		),
		formProof: newFormProofSigner(cfg.formHMACKey),
	}

	logger.Info("starting server", "addr", cfg.addr)
	// Call the new app.routes() method to get the servemux containing our routes,
	// and pass that to http.ListenAndServe().
	err = http.ListenAndServe(cfg.addr, app.routes())

	logger.Error(err.Error())
	os.Exit(1)
}

func loadLeadProtectionConfig(cfg *config) error {
	cfg.turnstileSiteKey = strings.TrimSpace(os.Getenv("TURNSTILE_SITE_KEY"))
	cfg.turnstileSecretKey = strings.TrimSpace(os.Getenv("TURNSTILE_SECRET_KEY"))
	cfg.turnstileVerifyURL = strings.TrimSpace(os.Getenv("TURNSTILE_VERIFY_URL"))
	cfg.turnstileAllowedHosts = parseAllowedHostnames(os.Getenv("TURNSTILE_ALLOWED_HOSTNAMES"))

	formKey := strings.TrimSpace(os.Getenv("FORM_HMAC_KEY"))
	if cfg.environment == "production" {
		if cfg.turnstileSiteKey == "" || cfg.turnstileSecretKey == "" {
			return fmt.Errorf("TURNSTILE_SITE_KEY and TURNSTILE_SECRET_KEY are required in production")
		}
		if cfg.turnstileSiteKey == testTurnstileSiteKey ||
			cfg.turnstileSiteKey == testTurnstileVisibleKey ||
			cfg.turnstileSecretKey == testTurnstileSecretKey {
			return fmt.Errorf("Turnstile test keys are not allowed in production")
		}
		if len(cfg.turnstileAllowedHosts) == 0 {
			return fmt.Errorf("TURNSTILE_ALLOWED_HOSTNAMES is required in production")
		}
		if formKey == "" {
			return fmt.Errorf("FORM_HMAC_KEY is required in production")
		}
		key, err := decodeFormHMACKey(formKey)
		if err != nil {
			return err
		}
		cfg.formHMACKey = key
		return nil
	}

	if cfg.turnstileSiteKey == "" {
		cfg.turnstileSiteKey = testTurnstileSiteKey
	}
	if cfg.turnstileSecretKey == "" {
		cfg.turnstileSecretKey = testTurnstileSecretKey
	}
	cfg.turnstileTestMode = cfg.turnstileSiteKey == testTurnstileSiteKey &&
		cfg.turnstileSecretKey == testTurnstileSecretKey
	if len(cfg.turnstileAllowedHosts) == 0 {
		cfg.turnstileAllowedHosts = map[string]struct{}{
			"localhost":      {},
			"127.0.0.1":      {},
			"dummy-key-pass": {},
		}
	}
	if formKey == "" {
		key := sha256.Sum256([]byte("tabiri-development-form-proof-key"))
		cfg.formHMACKey = key[:]
		return nil
	}
	key, err := decodeFormHMACKey(formKey)
	if err != nil {
		return err
	}
	cfg.formHMACKey = key
	return nil
}

func parseAllowedHostnames(value string) map[string]struct{} {
	hostnames := make(map[string]struct{})
	for _, hostname := range strings.Split(value, ",") {
		hostname = normalizeHostname(hostname)
		if hostname != "" {
			hostnames[hostname] = struct{}{}
		}
	}
	return hostnames
}
