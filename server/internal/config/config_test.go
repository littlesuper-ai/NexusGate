package config

import (
	"os"
	"testing"
)

func TestLoad_RequiresJWTSecret(t *testing.T) {
	// Ensure JWT_SECRET is not set
	os.Unsetenv("JWT_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is not set, got nil")
	}
	if err.Error() != "JWT_SECRET environment variable is required" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestLoad_DefaultValues(t *testing.T) {
	// Clear all env vars to test defaults
	envVars := []string{"LISTEN_ADDR", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE", "MQTT_BROKER", "CORS_ORIGINS"}
	saved := make(map[string]string)
	for _, key := range envVars {
		saved[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("JWT_SECRET", "test-secret-123")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":8080")
	}
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "localhost")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5432")
	}
	if cfg.DBUser != "nexusgate" {
		t.Errorf("DBUser = %q, want %q", cfg.DBUser, "nexusgate")
	}
	if cfg.DBPassword != "nexusgate" {
		t.Errorf("DBPassword = %q, want %q", cfg.DBPassword, "nexusgate")
	}
	if cfg.DBName != "nexusgate" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "nexusgate")
	}
	if cfg.DBSSLMode != "disable" {
		t.Errorf("DBSSLMode = %q, want %q", cfg.DBSSLMode, "disable")
	}
	if cfg.MQTTBroker != "tcp://localhost:1883" {
		t.Errorf("MQTTBroker = %q, want %q", cfg.MQTTBroker, "tcp://localhost:1883")
	}
	if cfg.JWTSecret != "test-secret-123" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "test-secret-123")
	}
}

func TestLoad_CORSOrigins(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name     string
		envVal   string
		expected []string
	}{
		{"empty", "", nil},
		{"single origin", "http://localhost:3000", []string{"http://localhost:3000"}},
		{"multiple origins", "http://a.com, http://b.com, http://c.com", []string{"http://a.com", "http://b.com", "http://c.com"}},
		{"with spaces", "  http://x.com ,  http://y.com  ", []string{"http://x.com", "http://y.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv("CORS_ORIGINS", tt.envVal)
			} else {
				os.Unsetenv("CORS_ORIGINS")
			}
			defer os.Unsetenv("CORS_ORIGINS")

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(cfg.CORSOrigins) != len(tt.expected) {
				t.Fatalf("CORSOrigins len = %d, want %d", len(cfg.CORSOrigins), len(tt.expected))
			}
			for i, want := range tt.expected {
				if cfg.CORSOrigins[i] != want {
					t.Errorf("CORSOrigins[%d] = %q, want %q", i, cfg.CORSOrigins[i], want)
				}
			}
		})
	}
}

func TestDSN(t *testing.T) {
	cfg := &Config{
		DBHost:     "db.example.com",
		DBPort:     "5433",
		DBUser:     "myuser",
		DBPassword: "mypass",
		DBName:     "mydb",
		DBSSLMode:  "require",
	}

	want := "host=db.example.com port=5433 user=myuser password=mypass dbname=mydb sslmode=require"
	got := cfg.DSN()
	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestLoad_CustomEnvValues(t *testing.T) {
	envs := map[string]string{
		"LISTEN_ADDR": ":9090",
		"DB_HOST":     "pghost",
		"DB_PORT":     "5433",
		"DB_USER":     "dbuser",
		"DB_PASSWORD":  "dbpass",
		"DB_NAME":     "dbname",
		"DB_SSLMODE":  "require",
		"MQTT_BROKER": "tcp://mqtt:1884",
		"JWT_SECRET":  "my-jwt-secret",
	}
	for k, v := range envs {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":9090")
	}
	if cfg.DBHost != "pghost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "pghost")
	}
	if cfg.MQTTBroker != "tcp://mqtt:1884" {
		t.Errorf("MQTTBroker = %q, want %q", cfg.MQTTBroker, "tcp://mqtt:1884")
	}
}
