package runtime

import (
	"errors"
	"time"
)

var (
	ErrTimelineQueryRequired = errors.New("timeline query requires session_id or task_id")
	ErrEvidenceQueryRequired = errors.New("evidence query requires action or request_id")
)

type TimelineEntry struct {
	TenantID    string
	SessionID   string
	TaskID      string
	EventType   string
	Status      string
	OccurredAt  time.Time
	RequestID   string
	ResourceRef string
}

type EvidenceEntry struct {
	TenantID    string
	SessionID   string
	TaskID      string
	EventType   string
	Action      string
	RequestID   string
	ResourceRef string
	OccurredAt  time.Time
}

type ReadSnapshot[T any] struct {
	Items      []T
	NextCursor string
	AsOf       time.Time
}
