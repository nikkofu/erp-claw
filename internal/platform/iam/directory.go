package iam

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	ErrActorNotFound = errors.New("actor not found")
	ErrInvalidActor  = errors.New("invalid actor")
)

// Directory stores actors and their role bindings per tenant.
type Directory interface {
	Save(ctx context.Context, tenantID string, actor Actor) error
	Get(ctx context.Context, tenantID, actorID string) (Actor, error)
	List(ctx context.Context, tenantID string) ([]Actor, error)
	Delete(ctx context.Context, tenantID, actorID string) error
}

type InMemoryDirectory struct {
	mu     sync.RWMutex
	actors map[string]Actor
}

func NewInMemoryDirectory() *InMemoryDirectory {
	return &InMemoryDirectory{
		actors: make(map[string]Actor),
	}
}

func (d *InMemoryDirectory) Save(_ context.Context, tenantID string, actor Actor) error {
	tenantID = strings.TrimSpace(tenantID)
	actor.ID = strings.TrimSpace(actor.ID)
	if tenantID == "" || actor.ID == "" {
		return ErrInvalidActor
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.actors[d.key(tenantID, actor.ID)] = cloneActor(actor)
	return nil
}

func (d *InMemoryDirectory) Get(_ context.Context, tenantID, actorID string) (Actor, error) {
	tenantID = strings.TrimSpace(tenantID)
	actorID = strings.TrimSpace(actorID)
	if tenantID == "" || actorID == "" {
		return Actor{}, ErrActorNotFound
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	actor, ok := d.actors[d.key(tenantID, actorID)]
	if !ok {
		return Actor{}, ErrActorNotFound
	}
	return cloneActor(actor), nil
}

func (d *InMemoryDirectory) List(_ context.Context, tenantID string) ([]Actor, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrActorNotFound
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make([]Actor, 0)
	prefix := tenantID + "/"
	for key, actor := range d.actors {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		out = append(out, cloneActor(actor))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (d *InMemoryDirectory) Delete(_ context.Context, tenantID, actorID string) error {
	tenantID = strings.TrimSpace(tenantID)
	actorID = strings.TrimSpace(actorID)
	if tenantID == "" || actorID == "" {
		return ErrActorNotFound
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	key := d.key(tenantID, actorID)
	if _, ok := d.actors[key]; !ok {
		return ErrActorNotFound
	}
	delete(d.actors, key)
	return nil
}

func (d *InMemoryDirectory) key(tenantID, actorID string) string {
	return tenantID + "/" + actorID
}

func cloneActor(actor Actor) Actor {
	actor.Roles = append([]string(nil), actor.Roles...)
	return actor
}
