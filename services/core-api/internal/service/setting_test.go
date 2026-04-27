package service

import (
	"context"
	"testing"
)

func TestSettingSeedDefaults(t *testing.T) {
	store := newFakeStore()
	svc := &Setting{q: store}
	if err := svc.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults() error = %v", err)
	}

	for key, want := range map[string]string{
		"default_max_attempts": "3",
		"max_concurrent":       "5",
		"headless":             "true",
		"scheduler_last_error": "",
	} {
		if got := store.settings[key]; got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}
