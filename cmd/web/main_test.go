package main

import (
	"encoding/base64"
	"testing"
)

func TestProductionLeadProtectionConfigurationIsRequired(t *testing.T) {
	t.Setenv("TURNSTILE_SITE_KEY", "")
	t.Setenv("TURNSTILE_SECRET_KEY", "")
	t.Setenv("TURNSTILE_ALLOWED_HOSTNAMES", "")
	t.Setenv("FORM_HMAC_KEY", "")

	cfg := config{environment: "production"}
	if err := loadLeadProtectionConfig(&cfg); err == nil {
		t.Fatal("expected missing production lead protection configuration to fail")
	}
}

func TestProductionLeadProtectionConfigurationLoads(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	t.Setenv("TURNSTILE_SITE_KEY", "site-key")
	t.Setenv("TURNSTILE_SECRET_KEY", "secret-key")
	t.Setenv("TURNSTILE_ALLOWED_HOSTNAMES", "www.tabirianalytics.com, tabirianalytics.com")
	t.Setenv("FORM_HMAC_KEY", key)

	cfg := config{environment: "production"}
	if err := loadLeadProtectionConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.turnstileAllowedHosts) != 2 {
		t.Fatalf("got %d allowed hostnames, want 2", len(cfg.turnstileAllowedHosts))
	}
	if len(cfg.formHMACKey) != 32 {
		t.Fatalf("got HMAC key length %d, want 32", len(cfg.formHMACKey))
	}
}

func TestProductionLeadProtectionConfigurationRejectsTestKeys(t *testing.T) {
	key := base64.RawURLEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	t.Setenv("TURNSTILE_SITE_KEY", testTurnstileSiteKey)
	t.Setenv("TURNSTILE_SECRET_KEY", testTurnstileSecretKey)
	t.Setenv("TURNSTILE_ALLOWED_HOSTNAMES", "www.tabirianalytics.com")
	t.Setenv("FORM_HMAC_KEY", key)

	cfg := config{environment: "production"}
	if err := loadLeadProtectionConfig(&cfg); err == nil {
		t.Fatal("expected production configuration to reject Turnstile test keys")
	}
}

func TestDevelopmentUsesInvisibleTurnstileTestKey(t *testing.T) {
	t.Setenv("TURNSTILE_SITE_KEY", "")
	t.Setenv("TURNSTILE_SECRET_KEY", "")
	t.Setenv("TURNSTILE_ALLOWED_HOSTNAMES", "")
	t.Setenv("FORM_HMAC_KEY", "")

	cfg := config{environment: "development"}
	if err := loadLeadProtectionConfig(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.turnstileSiteKey != testTurnstileSiteKey {
		t.Fatalf("got site key %q, want invisible test key", cfg.turnstileSiteKey)
	}
	if !cfg.turnstileTestMode {
		t.Fatal("expected Turnstile test mode to be enabled")
	}
}
