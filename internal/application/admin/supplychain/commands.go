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

type ListPurchaseOrdersInput struct {
	TenantID string
	Status   string
	Sort     string
	Page     int
	PageSize int
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

type ListApprovalRequestsInput struct {
	TenantID string
	Status   string
	Sort     string
	Page     int
	PageSize int
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

type ReserveInventoryInput struct {
	TenantID      string
	ActorID       string
	ProductID     string
	WarehouseID   string
	Quantity      int
	ReferenceType string
	ReferenceID   string
}

type IssueInventoryInput struct {
	TenantID      string
	ActorID       string
	ProductID     string
	WarehouseID   string
	Quantity      int
	ReferenceType string
	ReferenceID   string
}

type TransferInventoryInput struct {
	TenantID        string
	ActorID         string
	ProductID       string
	FromWarehouseID string
	ToWarehouseID   string
	Quantity        int
	ReferenceType   string
	ReferenceID     string
}

type CreateTransferOrderInput struct {
	TenantID        string
	ActorID         string
	ProductID       string
	FromWarehouseID string
	ToWarehouseID   string
	Quantity        int
}

type GetTransferOrderInput struct {
	TenantID        string
	TransferOrderID string
}

type ListTransferOrdersInput struct {
	TenantID string
	Status   string
	Sort     string
	Page     int
	PageSize int
}

type ExecuteTransferOrderInput struct {
	TenantID        string
	ActorID         string
	TransferOrderID string
}

type CancelTransferOrderInput struct {
	TenantID        string
	ActorID         string
	TransferOrderID string
}

type ListInventoryLedgerInput struct {
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

type ListPayableBillsInput struct {
	TenantID string
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

type CreateReceivableBillInput struct {
	TenantID    string
	ActorID     string
	ExternalRef string
}

type GetReceivableBillInput struct {
	TenantID string
	BillID   string
}

type ListReceivableBillsInput struct {
	TenantID string
}

type CreateSalesOrderLine struct {
	ProductID string
	Quantity  int
}

type CreateSalesOrderInput struct {
	TenantID    string
	ActorID     string
	WarehouseID string
	ExternalRef string
	Lines       []CreateSalesOrderLine
}

type GetSalesOrderInput struct {
	TenantID     string
	SalesOrderID string
}

type ListSalesOrdersInput struct {
	TenantID string
	Status   string
	Sort     string
	Page     int
	PageSize int
}

type ShipSalesOrderInput struct {
	TenantID     string
	ActorID      string
	SalesOrderID string
}

type GetBackofficeOverviewInput struct {
	TenantID string
}
