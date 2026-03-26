package memory

import (
	"context"
	"sync"

	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
)

type masterDataRepository struct {
	store *SupplyChainStore
}

type SupplyChainStore struct {
	mu         sync.RWMutex
	suppliers  map[string]masterdata.Supplier
	products   map[string]masterdata.Product
	warehouses map[string]masterdata.Warehouse
	orders     map[string]procurement.PurchaseOrder
	requests   map[string]approval.Request
}

func NewSupplyChainStore() *SupplyChainStore {
	return &SupplyChainStore{
		suppliers:  make(map[string]masterdata.Supplier),
		products:   make(map[string]masterdata.Product),
		warehouses: make(map[string]masterdata.Warehouse),
		orders:     make(map[string]procurement.PurchaseOrder),
		requests:   make(map[string]approval.Request),
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

func key(tenantID, id string) string {
	return tenantID + "/" + id
}

func clonePurchaseOrder(order procurement.PurchaseOrder) procurement.PurchaseOrder {
	order.Lines = append([]procurement.Line(nil), order.Lines...)
	return order
}
