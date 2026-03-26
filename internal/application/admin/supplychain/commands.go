package supplychain

type CreateSupplierInput struct {
	TenantID string
	ActorID  string
	Code     string
	Name     string
}

type CreateProductInput struct {
	TenantID string
	ActorID  string
	SKU      string
	Name     string
	Unit     string
}

type CreateWarehouseInput struct {
	TenantID string
	ActorID  string
	Code     string
	Name     string
}

type CreatePurchaseOrderLine struct {
	ProductID string
	Quantity  int
}

type CreatePurchaseOrderInput struct {
	TenantID    string
	ActorID     string
	SupplierID  string
	WarehouseID string
	Lines       []CreatePurchaseOrderLine
}

type SubmitPurchaseOrderInput struct {
	TenantID        string
	ActorID         string
	PurchaseOrderID string
}

type ResolveApprovalInput struct {
	TenantID   string
	ActorID    string
	ApprovalID string
}

type ReceivePurchaseOrderLine struct {
	ProductID string
	Quantity  int
}

type ReceivePurchaseOrderInput struct {
	TenantID        string
	ActorID         string
	PurchaseOrderID string
	Lines           []ReceivePurchaseOrderLine
}

type GetInventoryBalanceInput struct {
	TenantID    string
	ProductID   string
	WarehouseID string
}

type CreatePayableBillInput struct {
	TenantID        string
	ActorID         string
	PurchaseOrderID string
}

type GetPayableBillInput struct {
	TenantID string
	BillID   string
}

type CreatePayablePaymentPlanInput struct {
	TenantID       string
	ActorID        string
	PayableBillID  string
	DueDateISO8601 string
}

type ListPayablePaymentPlansInput struct {
	TenantID      string
	PayableBillID string
}
