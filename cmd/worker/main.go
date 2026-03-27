package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nikkofu/erp-claw/internal/bootstrap"
	messagingnats "github.com/nikkofu/erp-claw/internal/infrastructure/messaging/nats"
	"github.com/nikkofu/erp-claw/internal/infrastructure/persistence/postgres"
	"github.com/nikkofu/erp-claw/internal/platform/eventbus"
)

const (
	defaultOutboxBatchSize  = 100
	defaultOutboxRetryDelay = 5 * time.Second
	defaultOutboxMaxAttempt = 3
)

const claimPendingOutboxSQL = `
with picked as (
	select id
	from outbox
	where status = 'pending'
	  and available_at <= now()
	order by id
	for update skip locked
	limit $1
)
update outbox o
set status = 'processing'
from picked
where o.id = picked.id
returning o.id, o.tenant_id, o.topic, o.event_type, o.payload, o.attempts;
`

type outboxRecord struct {
	ID        int64
	TenantID  int64
	Topic     string
	EventType string
	Payload   []byte
	Attempts  int
}

type outboxStore interface {
	ClaimPending(ctx context.Context, limit int) ([]outboxRecord, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkPendingRetry(ctx context.Context, id int64, nextAvailableAt time.Time, attempts int, lastError string) error
	MarkFailed(ctx context.Context, id int64, attempts int, failedAt time.Time, lastError string) error
}

type postgresOutboxStore struct {
	db *sql.DB
}

func (s *postgresOutboxStore) ClaimPending(ctx context.Context, limit int) ([]outboxRecord, error) {
	rows, err := s.db.QueryContext(ctx, claimPendingOutboxSQL, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]outboxRecord, 0, limit)
	for rows.Next() {
		var rec outboxRecord
		if err := rows.Scan(&rec.ID, &rec.TenantID, &rec.Topic, &rec.EventType, &rec.Payload, &rec.Attempts); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *postgresOutboxStore) MarkPublished(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(
		ctx,
		`update outbox set status = 'published', published_at = now(), last_error = null where id = $1`,
		id,
	)
	return err
}

func (s *postgresOutboxStore) MarkPendingRetry(ctx context.Context, id int64, nextAvailableAt time.Time, attempts int, lastError string) error {
	_, err := s.db.ExecContext(
		ctx,
		`update outbox set status = 'pending', available_at = $2, published_at = null, attempts = $3, last_error = $4 where id = $1`,
		id,
		nextAvailableAt.UTC(),
		attempts,
		lastError,
	)
	return err
}

func (s *postgresOutboxStore) MarkFailed(ctx context.Context, id int64, attempts int, failedAt time.Time, lastError string) error {
	_, err := s.db.ExecContext(
		ctx,
		`update outbox set status = 'failed', attempts = $2, failed_at = $3, last_error = $4 where id = $1`,
		id,
		attempts,
		failedAt.UTC(),
		lastError,
	)
	return err
}

func main() {
	configPath := os.Getenv("ERP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/local/app.yaml"
	}

	cfg, err := bootstrap.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config (%s): %v", configPath, err)
	}

	bootstrap.StartRuntime(bootstrap.WorkerRole)
	shutdownTelemetry, err := bootstrap.SetupRuntimeTelemetry(cfg, bootstrap.WorkerRole)
	if err != nil {
		log.Fatalf("failed to setup telemetry: %v", err)
	}
	defer func() {
		if shutdownErr := shutdownTelemetry(context.Background()); shutdownErr != nil {
			log.Printf("failed to shutdown telemetry: %v", shutdownErr)
		}
	}()

	db, err := postgres.New(postgres.Config{
		DSN:          cfg.Database.DSN,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	})
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close postgres connection: %v", closeErr)
		}
	}()

	nc, err := messagingnats.New(messagingnats.Config{
		Servers: cfg.NATS.Servers,
		Cluster: cfg.NATS.Cluster,
	})
	if err != nil {
		log.Fatalf("failed to connect nats: %v", err)
	}
	defer nc.Close()

	bus, err := eventbus.NewNATS(nc)
	if err != nil {
		log.Fatalf("failed to initialize event bus: %v", err)
	}

	log.Printf("worker started (env=%s, config=%s, nats_servers=%v)", cfg.Env, configPath, cfg.NATS.Servers)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runOutboxPoller(ctx, db, bus, 2*time.Second); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}

func runOutboxPoller(ctx context.Context, db *sql.DB, bus eventbus.Bus, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Print("worker shutdown requested")
			return nil
		case <-ticker.C:
			if err := pollOutboxBatch(ctx, db, bus); err != nil {
				log.Printf("outbox poll iteration failed: %v", err)
			}
		}
	}
}

func pollOutboxBatch(ctx context.Context, db *sql.DB, bus eventbus.Bus) error {
	if db == nil {
		return fmt.Errorf("outbox db is required")
	}
	if bus == nil {
		return fmt.Errorf("event bus is required")
	}
	store := &postgresOutboxStore{db: db}
	return pollOutboxBatchWithStore(
		ctx,
		store,
		bus,
		time.Now(),
		defaultOutboxBatchSize,
		defaultOutboxRetryDelay,
		defaultOutboxMaxAttempt,
	)
}

func pollOutboxBatchWithStore(
	ctx context.Context,
	store outboxStore,
	bus eventbus.Bus,
	now time.Time,
	batchSize int,
	retryDelay time.Duration,
	maxAttempts int,
) error {
	if store == nil {
		return fmt.Errorf("outbox store is required")
	}
	if bus == nil {
		return fmt.Errorf("event bus is required")
	}
	if batchSize <= 0 {
		batchSize = defaultOutboxBatchSize
	}
	if retryDelay <= 0 {
		retryDelay = defaultOutboxRetryDelay
	}
	if maxAttempts <= 0 {
		maxAttempts = defaultOutboxMaxAttempt
	}

	records, err := store.ClaimPending(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("claim pending outbox: %w", err)
	}
	if len(records) == 0 {
		return nil
	}

	var firstErr error
	nextAvailableAt := now.UTC().Add(retryDelay)

	for _, rec := range records {
		evt := eventbus.Event{
			Topic:       rec.Topic,
			TenantID:    strconv.FormatInt(rec.TenantID, 10),
			Correlation: fmt.Sprintf("outbox:%d", rec.ID),
			Payload:     rec.Payload,
		}
		if err := bus.Publish(ctx, evt); err != nil {
			nextAttempts := rec.Attempts + 1
			log.Printf(
				"outbox publish failed (id=%d topic=%s event_type=%s attempts=%d): %v",
				rec.ID,
				rec.Topic,
				rec.EventType,
				nextAttempts,
				err,
			)
			lastError := err.Error()
			if nextAttempts >= maxAttempts {
				failedAt := now.UTC()
				if failErr := store.MarkFailed(ctx, rec.ID, nextAttempts, failedAt, lastError); failErr != nil {
					err = fmt.Errorf("publish failed: %v; mark failed failed: %w", err, failErr)
				}
				if firstErr == nil {
					firstErr = fmt.Errorf("publish outbox id=%d permanently failed: %w", rec.ID, err)
				}
				continue
			}
			if retryErr := store.MarkPendingRetry(ctx, rec.ID, nextAvailableAt, nextAttempts, lastError); retryErr != nil {
				err = fmt.Errorf("publish failed: %v; mark retry failed: %w", err, retryErr)
			}
			if firstErr == nil {
				firstErr = fmt.Errorf("publish outbox id=%d: %w", rec.ID, err)
			}
			continue
		}
		if err := store.MarkPublished(ctx, rec.ID); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("mark published outbox id=%d: %w", rec.ID, err)
			}
			continue
		}
	}

	if firstErr != nil {
		return firstErr
	}
	return nil
}
