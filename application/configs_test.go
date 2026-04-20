package application

import "testing"

func TestPostgresConfigValidate(t *testing.T) {
	t.Parallel()

	cfg := DefaultPostgresConfig()
	cfg.DSN = "postgres://user:pass@localhost:5432/db"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestPostgresConfigValidateMissingDSN(t *testing.T) {
	t.Parallel()

	cfg := DefaultPostgresConfig()
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing dsn")
	}
}

func TestRedisConfigValidate(t *testing.T) {
	t.Parallel()

	cfg := DefaultRedisConfig()
	cfg.Addr = "127.0.0.1:6379"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestS3ConfigValidate(t *testing.T) {
	t.Parallel()

	cfg := DefaultS3Config()
	cfg.Endpoint = "localhost:9000"
	cfg.AccessKey = "access"
	cfg.SecretKey = "secret"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}
