package config

import (
	"os"
	"testing"
)

func TestGetEnvReturnsFallback(t *testing.T) {
	key := "TEST_CONFIG_NOT_SET"
	_ = os.Unsetenv(key)

	got := getEnv(key, "fallback")

	if got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}
}

func TestGetEnvReturnsValue(t *testing.T) {
	key := "TEST_CONFIG_SET"
	t.Setenv(key, "value123")

	got := getEnv(key, "fallback")

	if got != "value123" {
		t.Fatalf("expected value123, got %q", got)
	}
}
