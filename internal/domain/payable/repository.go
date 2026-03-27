package payable

import (
	"context"
	"errors"
)

var ErrBillNotFound = errors.New("payable bill not found")
var ErrInvalidBillQuery = errors.New("invalid payable bill query")

type Repository interface {
	Save(ctx context.Context, bill Bill) error
	Get(ctx context.Context, tenantID, billID string) (Bill, error)
	GetByPurchaseOrder(ctx context.Context, tenantID, purchaseOrderID string) (Bill, error)
	ListByTenant(ctx context.Context, tenantID string) ([]Bill, error)
	SavePaymentPlan(ctx context.Context, plan PaymentPlan) error
	ListPaymentPlansByBill(ctx context.Context, tenantID, payableBillID string) ([]PaymentPlan, error)
}
