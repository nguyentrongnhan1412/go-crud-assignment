package config_test

import (
	"testing"

	"app/config"
)

func TestLoad_UsesDefaults(t *testing.T) {
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")

	cfg := config.Load()

	if cfg.ServerPort != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.ServerPort)
	}
	if cfg.DBHost != "localhost" {
		t.Fatalf("expected default host localhost, got %s", cfg.DBHost)
	}
	if cfg.DBName != "product_management" {
		t.Fatalf("expected default db name product_management, got %s", cfg.DBName)
	}
}

func TestLoad_UsesEnvironmentVariables(t *testing.T) {
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "appuser")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "products")
	t.Setenv("DB_SSLMODE", "require")

	cfg := config.Load()

	if cfg.ServerPort != "9090" || cfg.DBHost != "db" || cfg.DBPort != "5433" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.DBUser != "appuser" || cfg.DBPassword != "secret" || cfg.DBName != "products" {
		t.Fatalf("unexpected database config: %+v", cfg)
	}
	if cfg.DBSSLMode != "require" {
		t.Fatalf("expected sslmode require, got %s", cfg.DBSSLMode)
	}
}

func TestDatabaseDSN(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBName:     "product_management",
		DBSSLMode:  "disable",
	}

	dsn := cfg.DatabaseDSN()
	expected := "host=localhost port=5432 user=postgres password=postgres dbname=product_management sslmode=disable"
	if dsn != expected {
		t.Fatalf("expected dsn %q, got %q", expected, dsn)
	}
}
