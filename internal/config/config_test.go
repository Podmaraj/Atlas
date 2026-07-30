package config_test

import (
	"testing"
	"time"

	"edgecore/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("expected env to be 'development', got '%s'", cfg.Env)
	}

	if cfg.Gateway.HTTPPort != 8080 {
		t.Errorf("expected gateway HTTP port to be 8080, got %d", cfg.Gateway.HTTPPort)
	}

	if cfg.ControlPlane.Port != 8081 {
		t.Errorf("expected control plane port to be 8081, got %d", cfg.ControlPlane.Port)
	}

	if cfg.Redis.PubSubChannel != "edgecore:config:events" {
		t.Errorf("expected pubsub channel to be 'edgecore:config:events', got '%s'", cfg.Redis.PubSubChannel)
	}

	if cfg.Database.ConnMaxLifetime != 15*time.Minute {
		t.Errorf("expected conn max lifetime to be 15m, got %v", cfg.Database.ConnMaxLifetime)
	}
}
