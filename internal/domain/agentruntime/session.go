package agentruntime

import (
	"errors"
	"strings"
	"time"
)

type SessionStatus string

const (
	SessionStatusOpen   SessionStatus = "open"
	SessionStatusClosed SessionStatus = "closed"
)

var (
	errSessionTenantIDRequired        = errors.New("agent runtime session tenant id is required")
	errSessionKeyRequired             = errors.New("agent runtime session key is required")
	errSessionInvalidStatusTransition = errors.New("agent runtime session status transition is invalid")
)

type Session struct {
	ID         string
	TenantID   string
	SessionKey string
	Status     SessionStatus
	Metadata   map[string]any
	StartedAt  time.Time
	EndedAt    *time.Time
}

func NewSession(tenantID, sessionKey string, metadata map[string]any) (Session, error) {
	if strings.TrimSpace(tenantID) == "" {
		return Session{}, errSessionTenantIDRequired
	}
	if strings.TrimSpace(sessionKey) == "" {
		return Session{}, errSessionKeyRequired
	}

	return Session{
		TenantID:   tenantID,
		SessionKey: sessionKey,
		Status:     SessionStatusOpen,
		Metadata:   copyAnyMap(metadata),
	}, nil
}

func (s *Session) TransitionTo(next SessionStatus, endedAt time.Time) error {
	if !isValidSessionStatus(next) {
		return errSessionInvalidStatusTransition
	}
	if s.Status == next {
		return nil
	}

	if s.Status == SessionStatusOpen && next == SessionStatusClosed {
		s.Status = next
		if endedAt.IsZero() {
			endedAt = time.Now().UTC()
		}
		s.EndedAt = &endedAt
		return nil
	}

	return errSessionInvalidStatusTransition
}

func isValidSessionStatus(status SessionStatus) bool {
	switch status {
	case SessionStatusOpen, SessionStatusClosed:
		return true
	default:
		return false
	}
}
