package procurement

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPurchaseOrder          = errors.New("invalid purchase order")
	ErrPurchaseOrderAlreadySubmitted = errors.New("purchase order is not in draft status")
)

type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft           PurchaseOrderStatus = "draft"
	PurchaseOrderStatusPendingApproval PurchaseOrderStatus = "pending_approval"
	PurchaseOrderStatusApproved        PurchaseOrderStatus = "approved"
	PurchaseOrderStatusRejected        PurchaseOrderStatus = "rejected"
)

type Line struct {
	ProductID string
	Quantity  int
}

type PurchaseOrder struct {
	ID          string
	TenantID    string
	SupplierID  string
	WarehouseID string
	Status      PurchaseOrderStatus
	Lines       []Line
	ApprovalID  string
}

func NewPurchaseOrder(id, tenantID, supplierID, warehouseID string, lines []Line) (PurchaseOrder, error) {
	order := PurchaseOrder{
		ID:          strings.TrimSpace(id),
		TenantID:    strings.TrimSpace(tenantID),
		SupplierID:  strings.TrimSpace(supplierID),
		WarehouseID: strings.TrimSpace(warehouseID),
		Status:      PurchaseOrderStatusDraft,
		Lines:       append([]Line(nil), lines...),
	}
	if order.ID == "" || order.TenantID == "" || order.SupplierID == "" || order.WarehouseID == "" || len(order.Lines) == 0 {
		return PurchaseOrder{}, ErrInvalidPurchaseOrder
	}
	if err := validateLines(order.Lines); err != nil {
		return PurchaseOrder{}, err
	}
	return order, nil
}

func (po *PurchaseOrder) Submit(approvalID string) error {
	if po.Status != PurchaseOrderStatusDraft {
		return ErrPurchaseOrderAlreadySubmitted
	}
	if err := validateLines(po.Lines); err != nil {
		return err
	}
	approvalID = strings.TrimSpace(approvalID)
	if approvalID == "" {
		return ErrInvalidPurchaseOrder
	}
	po.Status = PurchaseOrderStatusPendingApproval
	po.ApprovalID = approvalID
	return nil
}

func (po *PurchaseOrder) MarkApproved() error {
	if po.Status != PurchaseOrderStatusPendingApproval {
		return ErrPurchaseOrderAlreadySubmitted
	}
	po.Status = PurchaseOrderStatusApproved
	return nil
}

func (po *PurchaseOrder) MarkRejected() error {
	if po.Status != PurchaseOrderStatusPendingApproval {
		return ErrPurchaseOrderAlreadySubmitted
	}
	po.Status = PurchaseOrderStatusRejected
	return nil
}

func validateLines(lines []Line) error {
	if len(lines) == 0 {
		return ErrInvalidPurchaseOrder
	}
	for _, line := range lines {
		if strings.TrimSpace(line.ProductID) == "" || line.Quantity <= 0 {
			return ErrInvalidPurchaseOrder
		}
	}
	return nil
}
