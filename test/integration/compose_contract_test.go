package integration

import (
	"os"
	"strings"
	"testing"
)

func TestDockerComposeIncludesRequiredServices(t *testing.T) {
	data, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("read compose: %v", err)
	}

	required := []string{"postgres", "redis", "nats", "minio", "otel-collector", "prometheus", "grafana"}
	content := string(data)
	for _, service := range required {
		if !strings.Contains(content, service+":") {
			t.Fatalf("expected service %q in compose file", service)
		}
	}
	if !strings.Contains(content, "configs/local/docker.env") {
		t.Fatalf("expected compose file to reference configs/local/docker.env so local env overrides work")
	}
}

func TestPhase2Wave1MigrationContract(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000002_init_phase2_wave1_tables.up.sql")
	if err != nil {
		t.Fatalf("read phase 2 migration: %v", err)
	}

	content := string(data)
	requiredTables := []string{
		"supplier",
		"product",
		"warehouse",
		"purchase_order",
		"purchase_order_line",
		"approval_request",
	}

	for _, table := range requiredTables {
		if !strings.Contains(content, "create table if not exists "+table) {
			t.Fatalf("expected migration to create table %q", table)
		}
	}

	requiredConstraints := []string{
		"unique (tenant_id, id)",
		"foreign key (tenant_id, supplier_id) references supplier(tenant_id, id)",
		"foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)",
		"foreign key (tenant_id, approval_id) references approval_request(tenant_id, id)",
		"foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id) on delete cascade",
		"foreign key (tenant_id, product_id) references product(tenant_id, id)",
	}
	for _, constraint := range requiredConstraints {
		if !strings.Contains(content, constraint) {
			t.Fatalf("expected migration to contain tenant-aware constraint %q", constraint)
		}
	}
}

func TestPhase2Wave2MigrationContract(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000003_init_phase2_wave2_inventory_tables.up.sql")
	if err != nil {
		t.Fatalf("read phase 2 wave 2 migration: %v", err)
	}

	content := string(data)
	requiredTables := []string{
		"receipt",
		"receipt_line",
		"inventory_ledger",
	}
	for _, table := range requiredTables {
		if !strings.Contains(content, "create table if not exists "+table) {
			t.Fatalf("expected migration to create table %q", table)
		}
	}

	requiredConstraints := []string{
		"foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id)",
		"foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)",
		"foreign key (tenant_id, receipt_id) references receipt(tenant_id, id) on delete cascade",
		"foreign key (tenant_id, product_id) references product(tenant_id, id)",
		"foreign key (tenant_id, reference_id) references receipt(tenant_id, id)",
	}
	for _, constraint := range requiredConstraints {
		if !strings.Contains(content, constraint) {
			t.Fatalf("expected migration to contain tenant-aware constraint %q", constraint)
		}
	}
}

func TestPhase2Wave3MigrationContract(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000004_init_phase2_wave3_payable_tables.up.sql")
	if err != nil {
		t.Fatalf("read phase 2 wave 3 migration: %v", err)
	}

	content := string(data)
	requiredTables := []string{
		"payable_bill",
	}
	for _, table := range requiredTables {
		if !strings.Contains(content, "create table if not exists "+table) {
			t.Fatalf("expected migration to create table %q", table)
		}
	}

	requiredConstraints := []string{
		"unique (tenant_id, id)",
		"unique (tenant_id, purchase_order_id)",
		"foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id)",
	}
	for _, constraint := range requiredConstraints {
		if !strings.Contains(content, constraint) {
			t.Fatalf("expected migration to contain tenant-aware constraint %q", constraint)
		}
	}
}

func TestPhase2Wave3PaymentPlanMigrationContract(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000005_init_phase2_wave3_payable_payment_plan_tables.up.sql")
	if err != nil {
		t.Fatalf("read phase 2 wave 3 payment plan migration: %v", err)
	}

	content := string(data)
	requiredTables := []string{
		"payable_payment_plan",
	}
	for _, table := range requiredTables {
		if !strings.Contains(content, "create table if not exists "+table) {
			t.Fatalf("expected migration to create table %q", table)
		}
	}

	requiredConstraints := []string{
		"unique (tenant_id, id)",
		"foreign key (tenant_id, payable_bill_id) references payable_bill(tenant_id, id) on delete cascade",
	}
	for _, constraint := range requiredConstraints {
		if !strings.Contains(content, constraint) {
			t.Fatalf("expected migration to contain tenant-aware constraint %q", constraint)
		}
	}
}

func TestPhase2Wave3ReceivableMigrationContract(t *testing.T) {
	data, err := os.ReadFile("../../migrations/000006_init_phase2_wave3_receivable_tables.up.sql")
	if err != nil {
		t.Fatalf("read phase 2 wave 3 receivable migration: %v", err)
	}

	content := string(data)
	requiredTables := []string{
		"receivable_bill",
	}
	for _, table := range requiredTables {
		if !strings.Contains(content, "create table if not exists "+table) {
			t.Fatalf("expected migration to create table %q", table)
		}
	}

	requiredConstraints := []string{
		"unique (tenant_id, id)",
		"unique (tenant_id, external_ref)",
	}
	for _, constraint := range requiredConstraints {
		if !strings.Contains(content, constraint) {
			t.Fatalf("expected migration to contain tenant-aware constraint %q", constraint)
		}
	}
}
