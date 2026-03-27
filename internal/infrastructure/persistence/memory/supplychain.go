package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
	"github.com/nikkofu/erp-claw/internal/domain/receivable"
	"github.com/nikkofu/erp-claw/internal/domain/sales"
)

type masterDataRepository struct {
	store *SupplyChainStore
}

type SupplyChainStore struct {
	mu             sync.RWMutex
	suppliers      map[string]masterdata.Supplier
	products       map[string]masterdata.Product
	warehouses     map[string]masterdata.Warehouse
	orders         map[string]procurement.PurchaseOrder
	requests       map[string]approval.Request
	receipts       map[string]inventory.Receipt
	ledger         map[string][]inventory.LedgerEntry
	bills          map[string]payable.Bill
	billsByPO      map[string]string
	plans          map[string]payable.PaymentPlan
	plansByBill    map[string][]string
	receivables    map[string]receivable.Bill
	reservations   map[string][]inventory.Reservation
	transferOrders map[string]inventory.TransferOrder
	salesOrders    map[string]sales.Order
}

func NewSupplyChainStore() *SupplyChainStore {
	return &SupplyChainStore{
		suppliers:      make(map[string]masterdata.Supplier),
		products:       make(map[string]masterdata.Product),
		warehouses:     make(map[string]masterdata.Warehouse),
		orders:         make(map[string]procurement.PurchaseOrder),
		requests:       make(map[string]approval.Request),
		receipts:       make(map[string]inventory.Receipt),
		ledger:         make(map[string][]inventory.LedgerEntry),
		bills:          make(map[string]payable.Bill),
		billsByPO:      make(map[string]string),
		plans:          make(map[string]payable.PaymentPlan),
		plansByBill:    make(map[string][]string),
		receivables:    make(map[string]receivable.Bill),
		reservations:   make(map[string][]inventory.Reservation),
		transferOrders: make(map[string]inventory.TransferOrder),
		salesOrders:    make(map[string]sales.Order),
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

func (r *inventoryRepository) SaveReservation(_ context.Context, reservation inventory.Reservation) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	k := inventoryKey(reservation.TenantID, reservation.ProductID, reservation.WarehouseID)
	r.store.reservations[k] = append(r.store.reservations[k], cloneReservation(reservation))
	return nil
}

func (r *inventoryRepository) ListReservations(_ context.Context, tenantID, productID, warehouseID string) ([]inventory.Reservation, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	stored := r.store.reservations[inventoryKey(tenantID, productID, warehouseID)]
	out := make([]inventory.Reservation, 0, len(stored))
	for _, reservation := range stored {
		out = append(out, cloneReservation(reservation))
	}
	return out, nil
}

func (r *inventoryRepository) SaveTransferOrder(_ context.Context, order inventory.TransferOrder) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.transferOrders[key(order.TenantID, order.ID)] = cloneTransferOrder(order)
	return nil
}

func (r *inventoryRepository) GetTransferOrder(_ context.Context, tenantID, transferOrderID string) (inventory.TransferOrder, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	order, ok := r.store.transferOrders[key(tenantID, transferOrderID)]
	if !ok {
		return inventory.TransferOrder{}, inventory.ErrTransferOrderNotFound
	}
	return cloneTransferOrder(order), nil
}

func (r *inventoryRepository) ListTransferOrdersByTenant(_ context.Context, tenantID string) ([]inventory.TransferOrder, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]inventory.TransferOrder, 0)
	prefix := tenantID + "/"
	for k, order := range r.store.transferOrders {
		if strings.HasPrefix(k, prefix) {
			out = append(out, cloneTransferOrder(order))
		}
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

func (r *payableRepository) ListByTenant(_ context.Context, tenantID string) ([]payable.Bill, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]payable.Bill, 0)
	prefix := tenantID + "/"
	for k, bill := range r.store.bills {
		if strings.HasPrefix(k, prefix) {
			out = append(out, clonePayableBill(bill))
		}
	}
	return out, nil
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

type receivableRepository struct {
	store *SupplyChainStore
}

func NewReceivableRepository() receivable.Repository {
	return NewSupplyChainStore().ReceivableRepository()
}

func (s *SupplyChainStore) ReceivableRepository() receivable.Repository {
	return &receivableRepository{store: s}
}

func (r *receivableRepository) Save(_ context.Context, bill receivable.Bill) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()

	r.store.receivables[key(bill.TenantID, bill.ID)] = cloneReceivableBill(bill)
	return nil
}

func (r *receivableRepository) Get(_ context.Context, tenantID, billID string) (receivable.Bill, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	bill, ok := r.store.receivables[key(tenantID, billID)]
	if !ok {
		return receivable.Bill{}, receivable.ErrBillNotFound
	}
	return cloneReceivableBill(bill), nil
}

func (r *receivableRepository) ListByTenant(_ context.Context, tenantID string) ([]receivable.Bill, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]receivable.Bill, 0)
	prefix := tenantID + "/"
	for k, bill := range r.store.receivables {
		if strings.HasPrefix(k, prefix) {
			out = append(out, cloneReceivableBill(bill))
		}
	}
	return out, nil
}

type salesOrderRepository struct {
	store *SupplyChainStore
}

func NewSalesOrderRepository() sales.Repository {
	return NewSupplyChainStore().SalesOrderRepository()
}

func (s *SupplyChainStore) SalesOrderRepository() sales.Repository {
	return &salesOrderRepository{store: s}
}

func (r *salesOrderRepository) Save(_ context.Context, order sales.Order) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.salesOrders[key(order.TenantID, order.ID)] = cloneSalesOrder(order)
	return nil
}

func (r *salesOrderRepository) Get(_ context.Context, tenantID, orderID string) (sales.Order, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	order, ok := r.store.salesOrders[key(tenantID, orderID)]
	if !ok {
		return sales.Order{}, sales.ErrOrderNotFound
	}
	return cloneSalesOrder(order), nil
}

func (r *salesOrderRepository) ListByTenant(_ context.Context, tenantID string) ([]sales.Order, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()

	out := make([]sales.Order, 0)
	prefix := tenantID + "/"
	for k, order := range r.store.salesOrders {
		if strings.HasPrefix(k, prefix) {
			out = append(out, cloneSalesOrder(order))
		}
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

func cloneReservation(reservation inventory.Reservation) inventory.Reservation {
	return reservation
}

func cloneTransferOrder(order inventory.TransferOrder) inventory.TransferOrder {
	return order
}

func clonePayableBill(bill payable.Bill) payable.Bill {
	return bill
}

func clonePayablePaymentPlan(plan payable.PaymentPlan) payable.PaymentPlan {
	return plan
}

func cloneReceivableBill(bill receivable.Bill) receivable.Bill {
	return bill
}

func cloneSalesOrder(order sales.Order) sales.Order {
	order.Lines = append([]sales.Line(nil), order.Lines...)
	return order
}

type supplyChainSnapshot struct {
	suppliers      map[string]masterdata.Supplier
	products       map[string]masterdata.Product
	warehouses     map[string]masterdata.Warehouse
	orders         map[string]procurement.PurchaseOrder
	requests       map[string]approval.Request
	receipts       map[string]inventory.Receipt
	ledger         map[string][]inventory.LedgerEntry
	bills          map[string]payable.Bill
	billsByPO      map[string]string
	plans          map[string]payable.PaymentPlan
	plansByBill    map[string][]string
	receivables    map[string]receivable.Bill
	reservations   map[string][]inventory.Reservation
	transferOrders map[string]inventory.TransferOrder
	salesOrders    map[string]sales.Order
}

func (s *SupplyChainStore) snapshot() supplyChainSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return supplyChainSnapshot{
		suppliers:      cloneSupplierMap(s.suppliers),
		products:       cloneProductMap(s.products),
		warehouses:     cloneWarehouseMap(s.warehouses),
		orders:         clonePurchaseOrderMap(s.orders),
		requests:       cloneApprovalRequestMap(s.requests),
		receipts:       cloneReceiptMap(s.receipts),
		ledger:         cloneLedgerMap(s.ledger),
		bills:          clonePayableBillMap(s.bills),
		billsByPO:      cloneStringMap(s.billsByPO),
		plans:          clonePayablePlanMap(s.plans),
		plansByBill:    cloneStringSliceMap(s.plansByBill),
		receivables:    cloneReceivableBillMap(s.receivables),
		reservations:   cloneReservationMap(s.reservations),
		transferOrders: cloneTransferOrderMap(s.transferOrders),
		salesOrders:    cloneSalesOrderMap(s.salesOrders),
	}
}

func (s *SupplyChainStore) restore(snapshot supplyChainSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.suppliers = snapshot.suppliers
	s.products = snapshot.products
	s.warehouses = snapshot.warehouses
	s.orders = snapshot.orders
	s.requests = snapshot.requests
	s.receipts = snapshot.receipts
	s.ledger = snapshot.ledger
	s.bills = snapshot.bills
	s.billsByPO = snapshot.billsByPO
	s.plans = snapshot.plans
	s.plansByBill = snapshot.plansByBill
	s.receivables = snapshot.receivables
	s.reservations = snapshot.reservations
	s.transferOrders = snapshot.transferOrders
	s.salesOrders = snapshot.salesOrders
}

func cloneSupplierMap(source map[string]masterdata.Supplier) map[string]masterdata.Supplier {
	out := make(map[string]masterdata.Supplier, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func cloneProductMap(source map[string]masterdata.Product) map[string]masterdata.Product {
	out := make(map[string]masterdata.Product, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func cloneWarehouseMap(source map[string]masterdata.Warehouse) map[string]masterdata.Warehouse {
	out := make(map[string]masterdata.Warehouse, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func clonePurchaseOrderMap(source map[string]procurement.PurchaseOrder) map[string]procurement.PurchaseOrder {
	out := make(map[string]procurement.PurchaseOrder, len(source))
	for k, v := range source {
		out[k] = clonePurchaseOrder(v)
	}
	return out
}

func cloneApprovalRequestMap(source map[string]approval.Request) map[string]approval.Request {
	out := make(map[string]approval.Request, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func cloneReceiptMap(source map[string]inventory.Receipt) map[string]inventory.Receipt {
	out := make(map[string]inventory.Receipt, len(source))
	for k, v := range source {
		out[k] = cloneReceipt(v)
	}
	return out
}

func cloneLedgerMap(source map[string][]inventory.LedgerEntry) map[string][]inventory.LedgerEntry {
	out := make(map[string][]inventory.LedgerEntry, len(source))
	for k, entries := range source {
		cloned := make([]inventory.LedgerEntry, 0, len(entries))
		for _, entry := range entries {
			cloned = append(cloned, cloneLedgerEntry(entry))
		}
		out[k] = cloned
	}
	return out
}

func clonePayableBillMap(source map[string]payable.Bill) map[string]payable.Bill {
	out := make(map[string]payable.Bill, len(source))
	for k, v := range source {
		out[k] = clonePayableBill(v)
	}
	return out
}

func cloneStringMap(source map[string]string) map[string]string {
	out := make(map[string]string, len(source))
	for k, v := range source {
		out[k] = v
	}
	return out
}

func clonePayablePlanMap(source map[string]payable.PaymentPlan) map[string]payable.PaymentPlan {
	out := make(map[string]payable.PaymentPlan, len(source))
	for k, v := range source {
		out[k] = clonePayablePaymentPlan(v)
	}
	return out
}

func cloneStringSliceMap(source map[string][]string) map[string][]string {
	out := make(map[string][]string, len(source))
	for k, values := range source {
		out[k] = append([]string(nil), values...)
	}
	return out
}

func cloneReceivableBillMap(source map[string]receivable.Bill) map[string]receivable.Bill {
	out := make(map[string]receivable.Bill, len(source))
	for k, v := range source {
		out[k] = cloneReceivableBill(v)
	}
	return out
}

func cloneReservationMap(source map[string][]inventory.Reservation) map[string][]inventory.Reservation {
	out := make(map[string][]inventory.Reservation, len(source))
	for k, reservations := range source {
		cloned := make([]inventory.Reservation, 0, len(reservations))
		for _, reservation := range reservations {
			cloned = append(cloned, cloneReservation(reservation))
		}
		out[k] = cloned
	}
	return out
}

func cloneTransferOrderMap(source map[string]inventory.TransferOrder) map[string]inventory.TransferOrder {
	out := make(map[string]inventory.TransferOrder, len(source))
	for k, v := range source {
		out[k] = cloneTransferOrder(v)
	}
	return out
}

func cloneSalesOrderMap(source map[string]sales.Order) map[string]sales.Order {
	out := make(map[string]sales.Order, len(source))
	for k, v := range source {
		out[k] = cloneSalesOrder(v)
	}
	return out
}
