package handler

import (
	"testing"
	"time"
)

func mustParseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

func TestMustParseTimeHelper(t *testing.T) {
	got := mustParseTime("2026-03-23T21:10:00+03:00")
	if got.IsZero() {
		t.Fatal("expected non-zero time")
	}
}
