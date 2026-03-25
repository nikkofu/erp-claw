package eventbus

import (
	"context"
	"sync"
)

type MemoryBus struct {
	mu     sync.Mutex
	events []Event
}

func NewMemory() *MemoryBus {
	return &MemoryBus{}
}

func (b *MemoryBus) Publish(_ context.Context, evt Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = append(b.events, evt)
	return nil
}
