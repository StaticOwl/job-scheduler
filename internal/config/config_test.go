package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("CHECK_INTERVAL", "30")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("CHECK_INTERVAL")
	}()

	cfg := Load()

	if cfg.DBHost != "testhost" {
		t.Errorf("Expected DBHost to be 'testhost', got '%s'", cfg.DBHost)
	}

	if cfg.DBPort != "5433" {
		t.Errorf("Expected DBPort to be '5433', got '%s'", cfg.DBPort)
	}

	if cfg.DBName != "testdb" {
		t.Errorf("Expected DBName to be 'testdb', got '%s'", cfg.DBName)
	}

	if cfg.DBUser != "testuser" {
		t.Errorf("Expected DBUser to be 'testuser', got '%s'", cfg.DBUser)
	}

	if cfg.DBPassword != "testpass" {
		t.Errorf("Expected DBPassword to be 'testpass', got '%s'", cfg.DBPassword)
	}

	if cfg.CheckInterval != 30 {
		t.Errorf("Expected CheckInterval to be 30, got %d", cfg.CheckInterval)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("CHECK_INTERVAL")

	cfg := Load()

	if cfg.DBHost != "localhost" {
		t.Errorf("Expected default DBHost to be 'localhost', got '%s'", cfg.DBHost)
	}

	if cfg.DBPort != "5432" {
		t.Errorf("Expected default DBPort to be '5432', got '%s'", cfg.DBPort)
	}

	if cfg.DBName != "scheduler_db" {
		t.Errorf("Expected default DBName to be 'scheduler_db', got '%s'", cfg.DBName)
	}

	if cfg.CheckInterval != 10 {
		t.Errorf("Expected default CheckInterval to be 10, got %d", cfg.CheckInterval)
	}
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "testdb",
		DBUser:     "testuser",
		DBPassword: "testpass",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	dsn := cfg.GetDSN()

	if dsn != expected {
		t.Errorf("Expected DSN to be '%s', got '%s'", expected, dsn)
	}
}

func TestCheckIntervalInvalidValue(t *testing.T) {
	os.Setenv("CHECK_INTERVAL", "invalid")
	defer os.Unsetenv("CHECK_INTERVAL")

	cfg := Load()

	if cfg.CheckInterval != 10 {
		t.Errorf("Expected CheckInterval to default to 10 on invalid value, got %d", cfg.CheckInterval)
	}
}
