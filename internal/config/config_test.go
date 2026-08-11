package config_test

import (
	"testing"

	"github.com/Mind2Screen-Dev-Team/POS-backend/internal/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	cfg := config.Load()

	if cfg.AppPort != "8080" {
		t.Errorf("expected AppPort 8080, got %s", cfg.AppPort)
	}
	if cfg.Addr() != "0.0.0.0:8080" {
		t.Errorf("expected Addr 0.0.0.0:8080, got %s", cfg.Addr())
	}
	if cfg.DSN() == "" {
		t.Error("expected non-empty DSN")
	}
}
