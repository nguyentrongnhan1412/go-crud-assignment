package config_test

import (
	"testing"
	"time"

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
	t.Setenv("REQUEST_TIMEOUT", "")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("DB_CONN_MAX_LIFETIME", "")

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
	if cfg.RequestTimeout != 30*time.Second {
		t.Fatalf("expected default request timeout 30s, got %s", cfg.RequestTimeout)
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Fatalf("expected default max open conns 25, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 5 {
		t.Fatalf("expected default max idle conns 5, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 5*time.Minute {
		t.Fatalf("expected default conn max lifetime 5m, got %s", cfg.DBConnMaxLifetime)
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

func TestLoad_UsesTimeoutAndPoolSettings(t *testing.T) {
	t.Setenv("REQUEST_TIMEOUT", "15s")
	t.Setenv("DB_MAX_OPEN_CONNS", "40")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME", "10m")

	cfg := config.Load()

	if cfg.RequestTimeout != 15*time.Second {
		t.Fatalf("expected request timeout 15s, got %s", cfg.RequestTimeout)
	}
	if cfg.DBMaxOpenConns != 40 {
		t.Fatalf("expected max open conns 40, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 10 {
		t.Fatalf("expected max idle conns 10, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 10*time.Minute {
		t.Fatalf("expected conn max lifetime 10m, got %s", cfg.DBConnMaxLifetime)
	}
}

func TestLoad_ClampsIdleConnsToOpenConns(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "5")
	t.Setenv("DB_MAX_IDLE_CONNS", "20")

	cfg := config.Load()

	if cfg.DBMaxOpenConns != 5 {
		t.Fatalf("expected max open conns 5, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 5 {
		t.Fatalf("expected max idle conns clamped to 5, got %d", cfg.DBMaxIdleConns)
	}
}

func TestLoad_IgnoresInvalidNumericAndDurationValues(t *testing.T) {
	t.Setenv("REQUEST_TIMEOUT", "not-a-duration")
	t.Setenv("DB_MAX_OPEN_CONNS", "abc")
	t.Setenv("DB_MAX_IDLE_CONNS", "0")
	t.Setenv("DB_CONN_MAX_LIFETIME", "-1m")

	cfg := config.Load()

	if cfg.RequestTimeout != 30*time.Second {
		t.Fatalf("expected fallback request timeout 30s, got %s", cfg.RequestTimeout)
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Fatalf("expected fallback max open conns 25, got %d", cfg.DBMaxOpenConns)
	}
	if cfg.DBMaxIdleConns != 5 {
		t.Fatalf("expected fallback max idle conns 5, got %d", cfg.DBMaxIdleConns)
	}
	if cfg.DBConnMaxLifetime != 5*time.Minute {
		t.Fatalf("expected fallback conn max lifetime 5m, got %s", cfg.DBConnMaxLifetime)
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
