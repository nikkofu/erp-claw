package masterdata

import (
	"errors"
	"strings"
)

var ErrInvalidProduct = errors.New("invalid product")

type Product struct {
	ID       string
	TenantID string
	SKU      string
	Name     string
	Unit     string
}

func NewProduct(id, tenantID, sku, name, unit string) (Product, error) {
	product := Product{
		ID:       strings.TrimSpace(id),
		TenantID: strings.TrimSpace(tenantID),
		SKU:      strings.TrimSpace(sku),
		Name:     strings.TrimSpace(name),
		Unit:     strings.TrimSpace(unit),
	}
	if product.ID == "" || product.TenantID == "" || product.SKU == "" || product.Name == "" || product.Unit == "" {
		return Product{}, ErrInvalidProduct
	}
	return product, nil
}
