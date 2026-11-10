package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"website.tabirianalytics.com/internal/database"
	sqlc "website.tabirianalytics.com/internal/database/sqlc"
)

func run(args []string, output io.Writer) (runErr error) {
	cfg, err := loadConfig(args)
	if err != nil {
		return err
	}
	logger := newLogger(output, cfg)

	db, err := database.Open(cfg.dbPath)
	if err != nil {
		return err
	}
	defer func() {
		runErr = errors.Join(runErr, db.Close())
	}()

	if err := database.ApplySchemaWithObserver(context.Background(), db, func(version int) {
		logger.Info(
			"database migration applied",
			"event", "database.migration.applied",
			"version", version,
		)
	}); err != nil {
		return err
	}
	templates, err := newTemplateCache()
	if err != nil {
		return err
	}
	telemetry, err := newTelemetry(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		runErr = errors.Join(runErr, telemetry.shutdown(shutdownCtx))
	}()

	app := &application{
		logger:  logger,
		cfg:     cfg,
		queries: sqlc.New(db),
		turnstile: newTurnstileHTTPVerifier(
			cfg.turnstileSecretKey,
			cfg.turnstileVerifyURL,
			cfg.turnstileAllowedHosts,
			cfg.turnstileTestMode,
		),
		formProof: newFormProofSigner(cfg.formHMACKey),
		templates: templates,
		telemetry: telemetry,
		now:       time.Now,
	}
	app.turnstile.(*turnstileHTTPVerifier).telemetry = telemetry

	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: cfg.readHeaderTimeout,
		ReadTimeout:       cfg.readTimeout,
		WriteTimeout:      cfg.writeTimeout,
		IdleTimeout:       cfg.idleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var workerDone <-chan struct{}
	stopWorker := func() {}
	if cfg.hubspotEnabled {
		workerCtx, cancelWorker := context.WithCancel(ctx)
		done := make(chan struct{})
		worker := &hubspotWorker{
			queries: app.queries,
			synchronizer: &hubspotClient{
				httpClient:               &http.Client{Timeout: cfg.hubspotRequestTimeout},
				baseURL:                  cfg.hubspotAPIBaseURL,
				serviceKey:               cfg.hubspotServiceKey,
				newsletterSubscriptionID: cfg.hubspotSubscriptionID,
			},
			logger:         logger,
			telemetry:      telemetry,
			pollInterval:   cfg.hubspotPollInterval,
			requestTimeout: cfg.hubspotRequestTimeout,
			now:            time.Now,
		}
		go func() {
			defer close(done)
			worker.run(workerCtx)
		}()
		workerDone = done
		stopWorker = cancelWorker
		logger.Info("HubSpot synchronization enabled")
	} else {
		logger.Info("HubSpot synchronization disabled")
	}

	err = serve(ctx, server, cfg, logger)
	stopWorker()
	if workerDone != nil {
		<-workerDone
	}
	return err
}

func serve(ctx context.Context, server *http.Server, cfg *config, logger *slog.Logger) error {
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("application started", "event", "application.started", "addr", cfg.addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		logger.Info("graceful shutdown started", "event", "application.shutdown.started")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("graceful shutdown completed", "event", "application.shutdown.completed")
	return nil
}

func newLogger(output io.Writer, cfg *config) *slog.Logger {
	level := new(slog.LevelVar)
	switch cfg.logLevel {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}
	options := &slog.HandlerOptions{Level: level}
	if cfg.environment == "production" {
		return slog.New(slog.NewJSONHandler(output, options))
	}
	return slog.New(slog.NewTextHandler(output, options))
}
