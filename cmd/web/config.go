package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type config struct {
	addr                  string
	staticDir             string
	dbPath                string
	environment           string
	logLevel              string
	readHeaderTimeout     time.Duration
	readTimeout           time.Duration
	writeTimeout          time.Duration
	idleTimeout           time.Duration
	shutdownTimeout       time.Duration
	telemetryEnabled      bool
	traceSampleRatio      float64
	turnstileSiteKey      string
	turnstileSecretKey    string
	turnstileVerifyURL    string
	turnstileAllowedHosts map[string]struct{}
	turnstileTestMode     bool
	formHMACKey           []byte
}

func loadConfig(args []string) (*config, error) {
	environment := strings.TrimSpace(os.Getenv("APP_ENV"))
	if environment == "" {
		environment = "development"
	}

	cfg := &config{}
	defaultTraceSampleRatio := 1.0
	if environment == "production" {
		defaultTraceSampleRatio = 0.1
	}
	if value := strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER_ARG")); value != "" {
		ratio, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("parse OTEL_TRACES_SAMPLER_ARG: %w", err)
		}
		defaultTraceSampleRatio = ratio
	}
	flags := flag.NewFlagSet("tabiri-web", flag.ContinueOnError)
	flags.StringVar(&cfg.addr, "addr", ":4000", "HTTP network address")
	flags.StringVar(&cfg.staticDir, "static-dir", "./ui/static", "Path to static assets")
	flags.StringVar(&cfg.dbPath, "db-path", "./tabiri.db", "Path to SQLite database")
	flags.StringVar(&cfg.environment, "environment", environment, "Application environment")
	flags.StringVar(&cfg.logLevel, "log-level", "info", "Log level: debug, info, warn, or error")
	flags.DurationVar(&cfg.readHeaderTimeout, "read-header-timeout", 5*time.Second, "HTTP header read timeout")
	flags.DurationVar(&cfg.readTimeout, "read-timeout", 15*time.Second, "HTTP request read timeout")
	flags.DurationVar(&cfg.writeTimeout, "write-timeout", 30*time.Second, "HTTP response write timeout")
	flags.DurationVar(&cfg.idleTimeout, "idle-timeout", 2*time.Minute, "HTTP keep-alive idle timeout")
	flags.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 10*time.Second, "Graceful shutdown timeout")
	flags.Float64Var(&cfg.traceSampleRatio, "trace-sample-ratio", defaultTraceSampleRatio, "Trace sampling ratio from 0 to 1")
	if err := flags.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	cfg.environment = strings.TrimSpace(cfg.environment)
	if cfg.environment == "" {
		return nil, fmt.Errorf("environment must not be empty")
	}
	if cfg.addr == "" {
		return nil, fmt.Errorf("addr must not be empty")
	}
	if cfg.staticDir == "" {
		return nil, fmt.Errorf("static-dir must not be empty")
	}
	if cfg.dbPath == "" {
		return nil, fmt.Errorf("db-path must not be empty")
	}
	if cfg.readHeaderTimeout <= 0 || cfg.readTimeout <= 0 || cfg.writeTimeout <= 0 ||
		cfg.idleTimeout <= 0 || cfg.shutdownTimeout <= 0 {
		return nil, fmt.Errorf("HTTP and shutdown timeouts must be positive")
	}
	if cfg.traceSampleRatio < 0 || cfg.traceSampleRatio > 1 {
		return nil, fmt.Errorf("trace-sample-ratio must be between 0 and 1")
	}
	switch cfg.logLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, fmt.Errorf("unsupported log level %q", cfg.logLevel)
	}

	if err := loadLeadProtectionConfig(cfg); err != nil {
		return nil, err
	}
	cfg.telemetryEnabled = telemetryConfigured()
	return cfg, nil
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
