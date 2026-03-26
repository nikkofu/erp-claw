package masterdata

import (
	"context"
	"errors"
)

var (
	ErrSupplierNotFound  = errors.New("supplier not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrWarehouseNotFound = errors.New("warehouse not found")
)

type Repository interface {
	SaveSupplier(ctx context.Context, supplier Supplier) error
	GetSupplier(ctx context.Context, tenantID, supplierID string) (Supplier, error)
	SaveProduct(ctx context.Context, product Product) error
	GetProduct(ctx context.Context, tenantID, productID string) (Product, error)
	SaveWarehouse(ctx context.Context, warehouse Warehouse) error
	GetWarehouse(ctx context.Context, tenantID, warehouseID string) (Warehouse, error)
}
