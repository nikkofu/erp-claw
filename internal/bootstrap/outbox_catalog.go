package bootstrap

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	sharedoutbox "github.com/nikkofu/erp-claw/internal/application/shared/outbox"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
)

func newOutboxCatalog(cfg Config) OutboxCatalog {
	if shouldUseInMemoryCatalogFallback(cfg) {
		return NewInMemoryOutboxCatalogForTest()
	}

	catalog, err := newPostgresOutboxCatalog(cfg.Database)
	if err == nil {
		return catalog
	}

	panic(fmt.Errorf("bootstrap: outbox catalog init failed: %w", err))
}

func newPostgresOutboxCatalog(cfg DatabaseConfig) (OutboxCatalog, error) {
	db, err := postgres.New(postgres.Config{
		DSN:          cfg.DSN,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo, err := postgres.NewOutboxRepository(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

type inMemoryOutboxCatalog struct {
	mu       sync.RWMutex
	messages map[int64]sharedoutbox.Message
}

func NewInMemoryOutboxCatalogForTest() *inMemoryOutboxCatalog {
	return &inMemoryOutboxCatalog{
		messages: make(map[int64]sharedoutbox.Message),
	}
}

func (c *inMemoryOutboxCatalog) StoreMessage(message sharedoutbox.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.AvailableAt.IsZero() {
		message.AvailableAt = now
	}
	if message.Status == "" {
		message.Status = "pending"
	}
	c.messages[message.ID] = message
}

func (c *inMemoryOutboxCatalog) ListMessages(_ context.Context, tenantID, status string, limit int) ([]sharedoutbox.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages := make([]sharedoutbox.Message, 0)
	for _, message := range c.messages {
		if tenantID != "" && message.TenantID != tenantID {
			continue
		}
		if status != "" && message.Status != status {
			continue
		}
		messages = append(messages, message)
	}

	sort.Slice(messages, func(i, j int) bool {
		if messages[i].CreatedAt.Equal(messages[j].CreatedAt) {
			return messages[i].ID > messages[j].ID
		}
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})
	if limit > 0 && len(messages) > limit {
		return messages[:limit], nil
	}
	return messages, nil
}

func (c *inMemoryOutboxCatalog) RequeueFailed(_ context.Context, ids []int64, availableAt time.Time) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	requeued := 0
	for _, id := range ids {
		message, ok := c.messages[id]
		if !ok || message.Status != "failed" {
			continue
		}
		message.Status = "pending"
		message.LastError = ""
		message.AvailableAt = availableAt
		message.ProcessingAt = nil
		c.messages[id] = message
		requeued++
	}
	return requeued, nil
}
