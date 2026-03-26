package ws

import (
	"context"
	"errors"
	"strings"
	"sync"

	platformruntime "github.com/nikkofu/erp-claw/internal/platform/runtime"
)

var errUnknownWorkspaceSession = errors.New("workspace session is not registered")

// WorkspaceGateway is the single runtime seam for workspace event fan-out.
// Full WebSocket protocol negotiation and agent task streaming are intentionally
// deferred to later plans; this type only preserves session-aware registration.
type WorkspaceGateway struct {
	mu       sync.RWMutex
	channels map[string]chan platformruntime.WorkspaceEvent
	history  map[string][]platformruntime.WorkspaceEvent
}

func NewWorkspaceGateway() *WorkspaceGateway {
	return &WorkspaceGateway{
		channels: make(map[string]chan platformruntime.WorkspaceEvent),
		history:  make(map[string][]platformruntime.WorkspaceEvent),
	}
}

func (g *WorkspaceGateway) RegisterChannel(sessionID string, buffer int) (<-chan platformruntime.WorkspaceEvent, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, errors.New("workspace session id is required")
	}
	if buffer <= 0 {
		buffer = 32
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	ch := make(chan platformruntime.WorkspaceEvent, buffer)
	g.channels[sessionID] = ch
	return ch, nil
}

func (g *WorkspaceGateway) UnregisterChannel(sessionID string) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if ch, ok := g.channels[sessionID]; ok {
		delete(g.channels, sessionID)
		close(ch)
	}
}

func (g *WorkspaceGateway) Broadcast(evt platformruntime.WorkspaceEvent) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if sessionID := strings.TrimSpace(evt.SessionID); sessionID != "" {
		ch, ok := g.channels[sessionID]
		if !ok {
			return errUnknownWorkspaceSession
		}
		select {
		case ch <- evt:
		default:
		}
		return nil
	}

	for _, ch := range g.channels {
		select {
		case ch <- evt:
		default:
		}
	}
	return nil
}

func (g *WorkspaceGateway) AppendWorkspaceEvent(_ context.Context, evt platformruntime.WorkspaceEvent) error {
	g.appendHistory(evt)
	return g.Broadcast(evt)
}

func (g *WorkspaceGateway) ChannelCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.channels)
}

func (g *WorkspaceGateway) ListWorkspaceEvents(_ context.Context, tenantID, sessionID string) ([]platformruntime.WorkspaceEvent, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, errors.New("workspace session id is required")
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	stored := g.history[sessionID]
	events := make([]platformruntime.WorkspaceEvent, 0, len(stored))
	for _, evt := range stored {
		if tenantID != "" && strings.TrimSpace(evt.TenantID) != strings.TrimSpace(tenantID) {
			continue
		}
		events = append(events, evt)
	}
	return events, nil
}

func (g *WorkspaceGateway) appendHistory(evt platformruntime.WorkspaceEvent) {
	sessionID := strings.TrimSpace(evt.SessionID)
	if sessionID == "" {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.history[sessionID] = append(g.history[sessionID], evt)
}
