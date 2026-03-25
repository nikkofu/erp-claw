package ws

import (
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
}

func NewWorkspaceGateway() *WorkspaceGateway {
	return &WorkspaceGateway{
		channels: make(map[string]chan platformruntime.WorkspaceEvent),
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

func (g *WorkspaceGateway) ChannelCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.channels)
}
