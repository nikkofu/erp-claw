package postgres

import (
	"testing"

	"github.com/nikkofu/erp-claw/internal/domain/approval"
)

var (
	_ approval.DefinitionRepository = (*ApprovalRepository)(nil)
	_ approval.InstanceRepository   = (*ApprovalRepository)(nil)
	_ approval.TaskRepository       = (*ApprovalRepository)(nil)
)

func TestNewApprovalRepositoryRejectsNilDB(t *testing.T) {
	_, err := NewApprovalRepository(nil)
	if err == nil {
		t.Fatal("expected nil db to fail")
	}
}
