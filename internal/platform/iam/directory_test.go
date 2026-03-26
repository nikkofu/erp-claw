package iam

import (
	"context"
	"errors"
	"testing"
)

func TestInMemoryDirectorySavesAndLoadsActor(t *testing.T) {
	directory := NewInMemoryDirectory()

	actor := Actor{
		ID:           "actor-a",
		Roles:        []string{"platform_admin", "supplychain_operator"},
		DepartmentID: "ops",
	}
	if err := directory.Save(context.Background(), "tenant-admin", actor); err != nil {
		t.Fatalf("save actor: %v", err)
	}

	got, err := directory.Get(context.Background(), "tenant-admin", "actor-a")
	if err != nil {
		t.Fatalf("get actor: %v", err)
	}
	if got.ID != "actor-a" {
		t.Fatalf("expected actor actor-a, got %s", got.ID)
	}
	if len(got.Roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(got.Roles))
	}
	if got.DepartmentID != "ops" {
		t.Fatalf("expected department ops, got %s", got.DepartmentID)
	}
}

func TestInMemoryDirectoryReturnsNotFoundForUnknownActor(t *testing.T) {
	directory := NewInMemoryDirectory()

	_, err := directory.Get(context.Background(), "tenant-admin", "missing")
	if !errors.Is(err, ErrActorNotFound) {
		t.Fatalf("expected ErrActorNotFound, got %v", err)
	}
}
