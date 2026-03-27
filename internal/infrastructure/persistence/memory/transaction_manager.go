package memory

import (
	"context"
	"sync"
)

// AtomicTransactionManager snapshots memory stores before command execution
// and restores snapshots when the command returns an error.
type AtomicTransactionManager struct {
	mu sync.Mutex

	supplyChain  *SupplyChainStore
	controlPlane *ControlPlaneStore
}

func NewAtomicTransactionManager(supplyChain *SupplyChainStore, controlPlane *ControlPlaneStore) *AtomicTransactionManager {
	return &AtomicTransactionManager{
		supplyChain:  supplyChain,
		controlPlane: controlPlane,
	}
}

func (m *AtomicTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if fn == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		supplySnapshot       supplyChainSnapshot
		controlPlaneSnapshot controlPlaneSnapshot
	)
	hasSupplyChain := m.supplyChain != nil
	hasControlPlane := m.controlPlane != nil
	if hasSupplyChain {
		supplySnapshot = m.supplyChain.snapshot()
	}
	if hasControlPlane {
		controlPlaneSnapshot = m.controlPlane.snapshot()
	}

	if err := fn(ctx); err != nil {
		if hasSupplyChain {
			m.supplyChain.restore(supplySnapshot)
		}
		if hasControlPlane {
			m.controlPlane.restore(controlPlaneSnapshot)
		}
		return err
	}
	return nil
}
