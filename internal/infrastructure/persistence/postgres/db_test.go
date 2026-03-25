package postgres

import "testing"

func TestConfigValidationRejectsEmptyDSN(t *testing.T) {
	t.Parallel()

	_, err := New(Config{})
	if err == nil {
		t.Fatal("expected error for empty DSN, got nil")
	}
}
