package bootstrap

import "testing"

func TestNewControlPlaneCatalogPanicsWhenDatabaseInitFailsOutsideTest(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected control-plane catalog bootstrap to panic on invalid runtime database")
		}
	}()

	_ = newControlPlaneCatalog(Config{
		Env: "local",
		Database: DatabaseConfig{
			DSN:          "postgres://invalid:invalid@127.0.0.1:1/erp_claw?sslmode=disable",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
}

func TestNewAgentRuntimeCatalogPanicsWhenDatabaseInitFailsOutsideTest(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected agent runtime catalog bootstrap to panic on invalid runtime database")
		}
	}()

	_ = newAgentRuntimeCatalog(Config{
		Env: "local",
		Database: DatabaseConfig{
			DSN:          "postgres://invalid:invalid@127.0.0.1:1/erp_claw?sslmode=disable",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
}

func TestNewApprovalCatalogPanicsWhenDatabaseInitFailsOutsideTest(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected approval catalog bootstrap to panic on invalid runtime database")
		}
	}()

	_ = newApprovalCatalog(Config{
		Env: "local",
		Database: DatabaseConfig{
			DSN:          "postgres://invalid:invalid@127.0.0.1:1/erp_claw?sslmode=disable",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
}

func TestNewCapabilityCatalogPanicsWhenDatabaseInitFailsOutsideTest(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected capability catalog bootstrap to panic on invalid runtime database")
		}
	}()

	_ = newCapabilityCatalog(Config{
		Env: "local",
		Database: DatabaseConfig{
			DSN:          "postgres://invalid:invalid@127.0.0.1:1/erp_claw?sslmode=disable",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
}

func TestNewGovernanceCatalogPanicsWhenDatabaseInitFailsOutsideTest(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected governance catalog bootstrap to panic on invalid runtime database")
		}
	}()

	_ = newGovernanceCatalog(Config{
		Env: "local",
		Database: DatabaseConfig{
			DSN:          "postgres://invalid:invalid@127.0.0.1:1/erp_claw?sslmode=disable",
			MaxOpenConns: 1,
			MaxIdleConns: 1,
		},
	})
}
