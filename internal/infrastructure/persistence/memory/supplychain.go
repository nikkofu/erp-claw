package memory

import (
	"context"
	"sync"

	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
)

type masterDataRepository struct {
	store *SupplyChainStore
}

type SupplyChainStore struct {
	mu          sync.RWMutex
	suppliers   map[string]masterdata.Supplier
	products    map[string]masterdata.Product
	warehouses  map[string]masterdata.Warehouse
	orders      map[string]procurement.PurchaseOrder
	requests    map[string]approval.Request
	receipts    map[string]inventory.Receipt
	ledger      map[string][]inventory.LedgerEntry
	bills       map[string]payable.Bill
	billsByPO   map[string]string
	plans       map[string]payable.PaymentPlan
	plansByBill map[string][]string
}

func NewSupplyChainStore() *SupplyChainStore {
	return &SupplyChainStore{
		suppliers:   make(map[string]masterdata.Supplier),
		products:    make(map[string]masterdata.Product),
		warehouses:  make(map[string]masterdata.Warehouse),
		orders:      make(map[string]procurement.PurchaseOrder),
		requests:    make(map[string]approval.Request),
		receipts:    make(map[string]inventory.Receipt),
		ledger:      make(map[string][]inventory.LedgerEntry),
		bills:       make(map[string]payable.Bill),
		billsByPO:   make(map[string]string),
		plans:       make(map[string]payable.PaymentPlan),
		plansByBill: make(map[string][]string),
	}
}

func NewMasterDataRepository() masterdata.Repository {
	return NewSupplyChainStore().MasterDataRepository()
}

func (s *SupplyChainStore) MasterDataRepository() masterdata.Repository {
	return &masterDataRepository{store: s}
}

func (r *masterDataRepository) SaveSupplier(_ context.Context, supplier masterdata.Supplier) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.suppliers[key(supplier.TenantID, supplier.ID)] = supplier
	return nil
}

func (r *masterDataRepository) GetSupplier(_ context.Context, tenantID, supplierID string) (masterdata.Supplier, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	supplier, ok := r.store.suppliers[key(tenantID, supplierID)]
	if !ok {
		return masterdata.Supplier{}, masterdata.ErrSupplierNotFound
	}
	return supplier, nil
}

func (r *masterDataRepository) SaveProduct(_ context.Context, product masterdata.Product) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.products[key(product.TenantID, product.ID)] = product
	return nil
}

func (r *masterDataRepository) GetProduct(_ context.Context, tenantID, productID string) (masterdata.Product, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	product, ok := r.store.products[key(tenantID, productID)]
	if !ok {
		return masterdata.Product{}, masterdata.ErrProductNotFound
	}
	return product, nil
}

func (r *masterDataRepository) SaveWarehouse(_ context.Context, warehouse masterdata.Warehouse) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.warehouses[key(warehouse.TenantID, warehouse.ID)] = warehouse
	return nil
}

func (r *masterDataRepository) GetWarehouse(_ context.Context, tenantID, warehouseID string) (masterdata.Warehouse, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	warehouse, ok := r.store.warehouses[key(tenantID, warehouseID)]
	if !ok {
		return masterdata.Warehouse{}, masterdata.ErrWarehouseNotFound
	}
	return warehouse, nil
}

type purchaseOrderRepository struct {
	store *SupplyChainStore
}

func NewPurchaseOrderRepository() procurement.Repository {
	return NewSupplyChainStore().PurchaseOrderRepository()
}

func (s *SupplyChainStore) PurchaseOrderRepository() procurement.Repository {
	return &purchaseOrderRepository{store: s}
}

func (r *purchaseOrderRepository) Save(_ context.Context, order procurement.PurchaseOrder) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.orders[key(order.TenantID, order.ID)] = clonePurchaseOrder(order)
	return nil
}

func (r *purchaseOrderRepository) Get(_ context.Context, tenantID, orderID string) (procurement.PurchaseOrder, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	order, ok := r.store.orders[key(tenantID, orderID)]
	if !ok {
		return procurement.PurchaseOrder{}, procurement.ErrPurchaseOrderNotFound
	}
	return clonePurchaseOrder(order), nil
}

type approvalRepository struct {
	store *SupplyChainStore
}

func NewApprovalRepository() approval.Repository {
	return NewSupplyChainStore().ApprovalRepository()
}

func (s *SupplyChainStore) ApprovalRepository() approval.Repository {
	return &approvalRepository{store: s}
}

func (r *approvalRepository) Save(_ context.Context, request approval.Request) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.requests[key(request.TenantID, request.ID)] = request
	return nil
}

func (r *approvalRepository) Get(_ context.Context, tenantID, requestID string) (approval.Request, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	request, ok := r.store.requests[key(tenantID, requestID)]
	if !ok {
		return approval.Request{}, approval.ErrRequestNotFound
	}
	return request, nil
}

type inventoryRepository struct {
	store *SupplyChainStore
}

func NewInventoryRepository() inventory.Repository {
	return NewSupplyChainStore().InventoryRepository()
}

func (s *SupplyChainStore) InventoryRepository() inventory.Repository {
	return &inventoryRepository{store: s}
}

func (r *inventoryRepository) SaveReceipt(_ context.Context, receipt inventory.Receipt) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.receipts[key(receipt.TenantID, receipt.ID)] = cloneReceipt(receipt)
	return nil
}

func (r *inventoryRepository) AppendLedgerEntries(_ context.Context, entries []inventory.LedgerEntry) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	for _, entry := range entries {
		k := inventoryKey(entry.TenantID, entry.ProductID, entry.WarehouseID)
		r.store.ledger[k] = append(r.store.ledger[k], cloneLedgerEntry(entry))
	}
	return nil
}

func (r *inventoryRepository) ListLedgerEntries(_ context.Context, tenantID, productID, warehouseID string) ([]inventory.LedgerEntry, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	stored := r.store.ledger[inventoryKey(tenantID, productID, warehouseID)]
	out := make([]inventory.LedgerEntry, 0, len(stored))
	for _, entry := range stored {
		out = append(out, cloneLedgerEntry(entry))
	}
	return out, nil
}

type payableRepository struct {
	store *SupplyChainStore
}

func NewPayableRepository() payable.Repository {
	return NewSupplyChainStore().PayableRepository()
}

func (s *SupplyChainStore) PayableRepository() payable.Repository {
	return &payableRepository{store: s}
}

func (r *payableRepository) Save(_ context.Context, bill payable.Bill) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	orderKey := key(bill.TenantID, bill.PurchaseOrderID)
	if existingID, exists := r.store.billsByPO[orderKey]; exists && existingID != bill.ID {
		return payable.ErrBillAlreadyExists
	}

	r.store.bills[key(bill.TenantID, bill.ID)] = clonePayableBill(bill)
	r.store.billsByPO[orderKey] = bill.ID
	return nil
}

func (r *payableRepository) Get(_ context.Context, tenantID, billID string) (payable.Bill, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	bill, ok := r.store.bills[key(tenantID, billID)]
	if !ok {
		return payable.Bill{}, payable.ErrBillNotFound
	}
	return clonePayableBill(bill), nil
}

func (r *payableRepository) GetByPurchaseOrder(_ context.Context, tenantID, purchaseOrderID string) (payable.Bill, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	billID, ok := r.store.billsByPO[key(tenantID, purchaseOrderID)]
	if !ok {
		return payable.Bill{}, payable.ErrBillNotFound
	}
	bill, ok := r.store.bills[key(tenantID, billID)]
	if !ok {
		return payable.Bill{}, payable.ErrBillNotFound
	}
	return clonePayableBill(bill), nil
}

func (r *payableRepository) SavePaymentPlan(_ context.Context, plan payable.PaymentPlan) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	if _, ok := r.store.bills[key(plan.TenantID, plan.PayableBillID)]; !ok {
		return payable.ErrBillNotFound
	}

	r.store.plans[key(plan.TenantID, plan.ID)] = clonePayablePaymentPlan(plan)
	billKey := key(plan.TenantID, plan.PayableBillID)
	r.store.plansByBill[billKey] = append(r.store.plansByBill[billKey], plan.ID)
	return nil
}

func (r *payableRepository) ListPaymentPlansByBill(_ context.Context, tenantID, payableBillID string) ([]payable.PaymentPlan, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	ids := r.store.plansByBill[key(tenantID, payableBillID)]
	out := make([]payable.PaymentPlan, 0, len(ids))
	for _, planID := range ids {
		plan, ok := r.store.plans[key(tenantID, planID)]
		if !ok {
			continue
		}
		out = append(out, clonePayablePaymentPlan(plan))
	}
	return out, nil
}

func key(tenantID, id string) string {
	return tenantID + "/" + id
}

func inventoryKey(tenantID, productID, warehouseID string) string {
	return tenantID + "/" + warehouseID + "/" + productID
}

func clonePurchaseOrder(order procurement.PurchaseOrder) procurement.PurchaseOrder {
	order.Lines = append([]procurement.Line(nil), order.Lines...)
	return order
}

func cloneReceipt(receipt inventory.Receipt) inventory.Receipt {
	receipt.Lines = append([]inventory.ReceiptLine(nil), receipt.Lines...)
	return receipt
}

func cloneLedgerEntry(entry inventory.LedgerEntry) inventory.LedgerEntry {
	return entry
}

func clonePayableBill(bill payable.Bill) payable.Bill {
	return bill
}

func clonePayablePaymentPlan(plan payable.PaymentPlan) payable.PaymentPlan {
	return plan
}
