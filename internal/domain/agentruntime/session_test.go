package agentruntime

import (
	"testing"
	"time"
)

func TestNewSessionRequiresTenantID(t *testing.T) {
	_, err := NewSession("", "session-a", map[string]any{"channel": "workspace"})
	if err == nil {
		t.Fatal("expected empty tenant id to fail")
	}
}

func TestSessionTransitionOpenToClosed(t *testing.T) {
	session, err := NewSession("1001", "session-a", map[string]any{"channel": "workspace"})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	endedAt := time.Now().UTC()
	if err := session.TransitionTo(SessionStatusClosed, endedAt); err != nil {
		t.Fatalf("transition to closed: %v", err)
	}

	if session.Status != SessionStatusClosed {
		t.Fatalf("expected closed status, got %q", session.Status)
	}
	if session.EndedAt == nil || !session.EndedAt.Equal(endedAt) {
		t.Fatalf("expected ended_at to be set to %s", endedAt)
	}
}

func TestSessionTransitionClosedToOpenRejected(t *testing.T) {
	session, err := NewSession("1001", "session-a", nil)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := session.TransitionTo(SessionStatusClosed, time.Now().UTC()); err != nil {
		t.Fatalf("transition to closed: %v", err)
	}

	err = session.TransitionTo(SessionStatusOpen, time.Now().UTC())
	if err == nil {
		t.Fatal("expected closed to open transition to fail")
	}
}
