package supplychain

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/inventory"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/payable"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
	"github.com/nikkofu/erp-claw/internal/domain/receivable"
)

type ServiceDeps struct {
	MasterData     masterdata.Repository
	PurchaseOrders procurement.Repository
	Approvals      approval.Repository
	Inventory      inventory.Repository
	Payables       payable.Repository
	Receivables    receivable.Repository
	Pipeline       *shared.Pipeline
}

type Service struct {
	masterData     masterdata.Repository
	purchaseOrders procurement.Repository
	approvals      approval.Repository
	inventory      inventory.Repository
	payables       payable.Repository
	receivables    receivable.Repository
	pipeline       *shared.Pipeline
}

var ids atomic.Uint64

func NewService(deps ServiceDeps) *Service {
	if deps.Pipeline == nil {
		deps.Pipeline = shared.NewPipeline(shared.PipelineDeps{})
	}
	return &Service{
		masterData:     deps.MasterData,
		purchaseOrders: deps.PurchaseOrders,
		approvals:      deps.Approvals,
		inventory:      deps.Inventory,
		payables:       deps.Payables,
		receivables:    deps.Receivables,
		pipeline:       deps.Pipeline,
	}
}

func (s *Service) CreateSupplier(ctx context.Context, input CreateSupplierInput) (masterdata.Supplier, error) {
	var supplier masterdata.Supplier
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "masterdata.suppliers.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		created, err := masterdata.NewSupplier(nextID("sup"), input.TenantID, input.Code, input.Name)
		if err != nil {
			return err
		}
		if err := s.masterData.SaveSupplier(txCtx, created); err != nil {
			return err
		}
		supplier = created
		return nil
	})
	return supplier, err
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (masterdata.Product, error) {
	var product masterdata.Product
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "masterdata.products.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		created, err := masterdata.NewProduct(nextID("prd"), input.TenantID, input.SKU, input.Name, input.Unit)
		if err != nil {
			return err
		}
		if err := s.masterData.SaveProduct(txCtx, created); err != nil {
			return err
		}
		product = created
		return nil
	})
	return product, err
}

func (s *Service) CreateWarehouse(ctx context.Context, input CreateWarehouseInput) (masterdata.Warehouse, error) {
	var warehouse masterdata.Warehouse
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "masterdata.warehouses.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		created, err := masterdata.NewWarehouse(nextID("wh"), input.TenantID, input.Code, input.Name)
		if err != nil {
			return err
		}
		if err := s.masterData.SaveWarehouse(txCtx, created); err != nil {
			return err
		}
		warehouse = created
		return nil
	})
	return warehouse, err
}

func (s *Service) CreatePurchaseOrder(ctx context.Context, input CreatePurchaseOrderInput) (procurement.PurchaseOrder, error) {
	var order procurement.PurchaseOrder
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "procurement.purchase_orders.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if _, err := s.masterData.GetSupplier(txCtx, input.TenantID, input.SupplierID); err != nil {
			return err
		}
		if _, err := s.masterData.GetWarehouse(txCtx, input.TenantID, input.WarehouseID); err != nil {
			return err
		}

		lines := make([]procurement.Line, 0, len(input.Lines))
		for _, line := range input.Lines {
			if _, err := s.masterData.GetProduct(txCtx, input.TenantID, line.ProductID); err != nil {
				return err
			}
			lines = append(lines, procurement.Line{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
			})
		}

		created, err := procurement.NewPurchaseOrder(nextID("po"), input.TenantID, input.SupplierID, input.WarehouseID, lines)
		if err != nil {
			return err
		}
		if err := s.purchaseOrders.Save(txCtx, created); err != nil {
			return err
		}
		order = created
		return nil
	})
	return order, err
}

func (s *Service) SubmitPurchaseOrder(ctx context.Context, input SubmitPurchaseOrderInput) (procurement.PurchaseOrder, approval.Request, error) {
	var (
		order   procurement.PurchaseOrder
		request approval.Request
	)
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "procurement.purchase_orders.submit",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		current, err := s.purchaseOrders.Get(txCtx, input.TenantID, input.PurchaseOrderID)
		if err != nil {
			return err
		}

		createdRequest, err := approval.NewRequest(nextID("apr"), input.TenantID, "purchase_order", current.ID, input.ActorID)
		if err != nil {
			return err
		}
		if err := current.Submit(createdRequest.ID); err != nil {
			return err
		}

		if err := s.purchaseOrders.Save(txCtx, current); err != nil {
			return err
		}
		if err := s.approvals.Save(txCtx, createdRequest); err != nil {
			return err
		}

		order = current
		request = createdRequest
		return nil
	})
	return order, request, err
}

func (s *Service) ApproveRequest(ctx context.Context, input ResolveApprovalInput) (procurement.PurchaseOrder, approval.Request, error) {
	return s.resolveRequest(ctx, input, func(request *approval.Request) error {
		return request.Approve(input.ActorID)
	}, func(order *procurement.PurchaseOrder) error {
		return order.MarkApproved()
	}, "approval.requests.approve")
}

func (s *Service) RejectRequest(ctx context.Context, input ResolveApprovalInput) (procurement.PurchaseOrder, approval.Request, error) {
	return s.resolveRequest(ctx, input, func(request *approval.Request) error {
		return request.Reject(input.ActorID)
	}, func(order *procurement.PurchaseOrder) error {
		return order.MarkRejected()
	}, "approval.requests.reject")
}

func (s *Service) GetPurchaseOrder(ctx context.Context, tenantID, orderID string) (procurement.PurchaseOrder, approval.Request, error) {
	order, err := s.purchaseOrders.Get(ctx, tenantID, orderID)
	if err != nil {
		return procurement.PurchaseOrder{}, approval.Request{}, err
	}
	if order.ApprovalID == "" {
		return order, approval.Request{}, nil
	}

	request, err := s.approvals.Get(ctx, tenantID, order.ApprovalID)
	if err != nil {
		return procurement.PurchaseOrder{}, approval.Request{}, err
	}
	return order, request, nil
}

func (s *Service) ReceivePurchaseOrder(ctx context.Context, input ReceivePurchaseOrderInput) (inventory.Receipt, []inventory.LedgerEntry, procurement.PurchaseOrder, error) {
	var (
		receipt       inventory.Receipt
		ledgerEntries []inventory.LedgerEntry
		order         procurement.PurchaseOrder
	)
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "procurement.purchase_orders.receive",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		currentOrder, err := s.purchaseOrders.Get(txCtx, input.TenantID, input.PurchaseOrderID)
		if err != nil {
			return err
		}

		receiptLines := make([]inventory.ReceiptLine, 0, len(input.Lines))
		for _, line := range input.Lines {
			receiptLines = append(receiptLines, inventory.ReceiptLine{
				ProductID: line.ProductID,
				Quantity:  line.Quantity,
			})
		}
		if err := validateReceiptLinesAgainstOrder(receiptLines, currentOrder.Lines); err != nil {
			return err
		}

		createdReceipt, err := inventory.NewReceipt(nextID("rcv"), input.TenantID, currentOrder.ID, currentOrder.WarehouseID, input.ActorID, receiptLines)
		if err != nil {
			return err
		}

		createdEntries := make([]inventory.LedgerEntry, 0, len(receiptLines))
		for _, line := range receiptLines {
			entry, err := inventory.NewInboundLedgerEntry(nextID("led"), input.TenantID, line.ProductID, currentOrder.WarehouseID, "receipt", createdReceipt.ID, line.Quantity)
			if err != nil {
				return err
			}
			createdEntries = append(createdEntries, entry)
		}

		if err := currentOrder.MarkReceived(); err != nil {
			return err
		}
		if err := s.purchaseOrders.Save(txCtx, currentOrder); err != nil {
			return err
		}
		if err := s.inventory.SaveReceipt(txCtx, createdReceipt); err != nil {
			return err
		}
		if err := s.inventory.AppendLedgerEntries(txCtx, createdEntries); err != nil {
			return err
		}

		receipt = createdReceipt
		ledgerEntries = createdEntries
		order = currentOrder
		return nil
	})
	return receipt, ledgerEntries, order, err
}

func (s *Service) GetInventoryBalance(ctx context.Context, input GetInventoryBalanceInput) (inventory.Balance, error) {
	entries, err := s.inventory.ListLedgerEntries(ctx, input.TenantID, input.ProductID, input.WarehouseID)
	if err != nil {
		return inventory.Balance{}, err
	}
	reservations, err := s.inventory.ListReservations(ctx, input.TenantID, input.ProductID, input.WarehouseID)
	if err != nil {
		return inventory.Balance{}, err
	}
	balance := inventory.Balance{
		TenantID:    input.TenantID,
		ProductID:   input.ProductID,
		WarehouseID: input.WarehouseID,
	}
	for _, entry := range entries {
		balance.OnHand += entry.QuantityDelta
	}
	for _, reservation := range reservations {
		if reservation.Status != inventory.ReservationStatusActive {
			continue
		}
		balance.Reserved += reservation.Quantity
	}
	balance.Available = balance.OnHand - balance.Reserved
	return balance, nil
}

func (s *Service) ReserveInventory(ctx context.Context, input ReserveInventoryInput) (inventory.Reservation, error) {
	var reservation inventory.Reservation
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "inventory.reservations.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if _, err := s.masterData.GetProduct(txCtx, input.TenantID, input.ProductID); err != nil {
			return err
		}
		if _, err := s.masterData.GetWarehouse(txCtx, input.TenantID, input.WarehouseID); err != nil {
			return err
		}

		balance, err := s.GetInventoryBalance(txCtx, GetInventoryBalanceInput{
			TenantID:    input.TenantID,
			ProductID:   input.ProductID,
			WarehouseID: input.WarehouseID,
		})
		if err != nil {
			return err
		}
		if input.Quantity > balance.Available {
			return inventory.ErrInsufficientAvailableInventory
		}

		created, err := inventory.NewReservation(
			nextID("rsv"),
			input.TenantID,
			input.ProductID,
			input.WarehouseID,
			input.ReferenceType,
			input.ReferenceID,
			input.ActorID,
			input.Quantity,
		)
		if err != nil {
			return err
		}
		if err := s.inventory.SaveReservation(txCtx, created); err != nil {
			return err
		}
		reservation = created
		return nil
	})
	return reservation, err
}

func (s *Service) IssueInventory(ctx context.Context, input IssueInventoryInput) (inventory.LedgerEntry, error) {
	var outbound inventory.LedgerEntry
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "inventory.outbounds.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if _, err := s.masterData.GetProduct(txCtx, input.TenantID, input.ProductID); err != nil {
			return err
		}
		if _, err := s.masterData.GetWarehouse(txCtx, input.TenantID, input.WarehouseID); err != nil {
			return err
		}

		balance, err := s.GetInventoryBalance(txCtx, GetInventoryBalanceInput{
			TenantID:    input.TenantID,
			ProductID:   input.ProductID,
			WarehouseID: input.WarehouseID,
		})
		if err != nil {
			return err
		}
		if input.Quantity > balance.Available {
			return inventory.ErrInsufficientAvailableInventory
		}

		entry, err := inventory.NewOutboundLedgerEntry(
			nextID("led"),
			input.TenantID,
			input.ProductID,
			input.WarehouseID,
			input.ReferenceType,
			input.ReferenceID,
			input.Quantity,
		)
		if err != nil {
			return err
		}
		if err := s.inventory.AppendLedgerEntries(txCtx, []inventory.LedgerEntry{entry}); err != nil {
			return err
		}
		outbound = entry
		return nil
	})
	return outbound, err
}

func (s *Service) TransferInventory(ctx context.Context, input TransferInventoryInput) ([]inventory.LedgerEntry, error) {
	var entries []inventory.LedgerEntry
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "inventory.transfers.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if _, err := s.masterData.GetProduct(txCtx, input.TenantID, input.ProductID); err != nil {
			return err
		}
		if _, err := s.masterData.GetWarehouse(txCtx, input.TenantID, input.FromWarehouseID); err != nil {
			return err
		}
		if _, err := s.masterData.GetWarehouse(txCtx, input.TenantID, input.ToWarehouseID); err != nil {
			return err
		}

		sourceBalance, err := s.GetInventoryBalance(txCtx, GetInventoryBalanceInput{
			TenantID:    input.TenantID,
			ProductID:   input.ProductID,
			WarehouseID: input.FromWarehouseID,
		})
		if err != nil {
			return err
		}
		if input.Quantity > sourceBalance.Available {
			return inventory.ErrInsufficientAvailableInventory
		}

		outbound, err := inventory.NewOutboundLedgerEntry(
			nextID("led"),
			input.TenantID,
			input.ProductID,
			input.FromWarehouseID,
			input.ReferenceType,
			input.ReferenceID,
			input.Quantity,
		)
		if err != nil {
			return err
		}
		inbound, err := inventory.NewInboundLedgerEntry(
			nextID("led"),
			input.TenantID,
			input.ProductID,
			input.ToWarehouseID,
			input.ReferenceType,
			input.ReferenceID,
			input.Quantity,
		)
		if err != nil {
			return err
		}
		created := []inventory.LedgerEntry{outbound, inbound}
		if err := s.inventory.AppendLedgerEntries(txCtx, created); err != nil {
			return err
		}
		entries = created
		return nil
	})
	return entries, err
}

func (s *Service) CreatePayableBill(ctx context.Context, input CreatePayableBillInput) (payable.Bill, error) {
	var bill payable.Bill
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "payable.bills.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		order, err := s.purchaseOrders.Get(txCtx, input.TenantID, input.PurchaseOrderID)
		if err != nil {
			return err
		}
		if order.Status != procurement.PurchaseOrderStatusReceived {
			return payable.ErrOrderNotBillable
		}
		if _, err := s.payables.GetByPurchaseOrder(txCtx, input.TenantID, order.ID); err == nil {
			return payable.ErrBillAlreadyExists
		} else if !errors.Is(err, payable.ErrBillNotFound) {
			return err
		}

		created, err := payable.NewBill(nextID("pab"), input.TenantID, order.ID, input.ActorID)
		if err != nil {
			return err
		}
		if err := s.payables.Save(txCtx, created); err != nil {
			return err
		}

		bill = created
		return nil
	})
	return bill, err
}

func (s *Service) GetPayableBill(ctx context.Context, input GetPayableBillInput) (payable.Bill, error) {
	return s.payables.Get(ctx, input.TenantID, input.BillID)
}

func (s *Service) ListPayableBills(ctx context.Context, input ListPayableBillsInput) ([]payable.Bill, error) {
	return s.payables.ListByTenant(ctx, input.TenantID)
}

func (s *Service) CreatePayablePaymentPlan(ctx context.Context, input CreatePayablePaymentPlanInput) (payable.PaymentPlan, error) {
	var plan payable.PaymentPlan
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "payable.payment_plans.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		if _, err := s.payables.Get(txCtx, input.TenantID, input.PayableBillID); err != nil {
			return err
		}
		created, err := payable.NewPaymentPlan(nextID("ppm"), input.TenantID, input.PayableBillID, input.ActorID, input.DueDateISO8601)
		if err != nil {
			return err
		}
		if err := s.payables.SavePaymentPlan(txCtx, created); err != nil {
			return err
		}
		plan = created
		return nil
	})
	return plan, err
}

func (s *Service) ListPayablePaymentPlans(ctx context.Context, input ListPayablePaymentPlansInput) ([]payable.PaymentPlan, error) {
	return s.payables.ListPaymentPlansByBill(ctx, input.TenantID, input.PayableBillID)
}

func (s *Service) CreateReceivableBill(ctx context.Context, input CreateReceivableBillInput) (receivable.Bill, error) {
	var bill receivable.Bill
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     "receivable.bills.create",
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		created, err := receivable.NewBill(nextID("reb"), input.TenantID, input.ExternalRef, input.ActorID)
		if err != nil {
			return err
		}
		if err := s.receivables.Save(txCtx, created); err != nil {
			return err
		}
		bill = created
		return nil
	})
	return bill, err
}

func (s *Service) GetReceivableBill(ctx context.Context, input GetReceivableBillInput) (receivable.Bill, error) {
	return s.receivables.Get(ctx, input.TenantID, input.BillID)
}

func (s *Service) ListReceivableBills(ctx context.Context, input ListReceivableBillsInput) ([]receivable.Bill, error) {
	return s.receivables.ListByTenant(ctx, input.TenantID)
}

func (s *Service) resolveRequest(
	ctx context.Context,
	input ResolveApprovalInput,
	resolveApproval func(*approval.Request) error,
	resolveOrder func(*procurement.PurchaseOrder) error,
	commandName string,
) (procurement.PurchaseOrder, approval.Request, error) {
	var (
		order   procurement.PurchaseOrder
		request approval.Request
	)
	err := s.pipeline.Execute(ctx, shared.Command{
		Name:     commandName,
		TenantID: input.TenantID,
		ActorID:  input.ActorID,
		Payload:  input,
	}, func(txCtx context.Context, _ shared.Command) error {
		currentRequest, err := s.approvals.Get(txCtx, input.TenantID, input.ApprovalID)
		if err != nil {
			return err
		}
		currentOrder, err := s.purchaseOrders.Get(txCtx, input.TenantID, currentRequest.ResourceID)
		if err != nil {
			return err
		}
		if err := resolveApproval(&currentRequest); err != nil {
			return err
		}
		if err := resolveOrder(&currentOrder); err != nil {
			return err
		}
		if err := s.approvals.Save(txCtx, currentRequest); err != nil {
			return err
		}
		if err := s.purchaseOrders.Save(txCtx, currentOrder); err != nil {
			return err
		}

		order = currentOrder
		request = currentRequest
		return nil
	})
	return order, request, err
}

func nextID(prefix string) string {
	return fmt.Sprintf("%s-%06d", prefix, ids.Add(1))
}

func validateReceiptLinesAgainstOrder(receiptLines []inventory.ReceiptLine, orderLines []procurement.Line) error {
	if _, err := inventory.NewReceipt("receipt-validation", "tenant-validation", "po-validation", "warehouse-validation", "actor-validation", receiptLines); err != nil {
		return err
	}
	if len(receiptLines) != len(orderLines) {
		return inventory.ErrInvalidReceipt
	}

	expected := make(map[string]int, len(orderLines))
	for _, line := range orderLines {
		expected[line.ProductID] += line.Quantity
	}
	for _, line := range receiptLines {
		if expected[line.ProductID] != line.Quantity {
			return inventory.ErrInvalidReceipt
		}
		delete(expected, line.ProductID)
	}
	if len(expected) > 0 {
		return inventory.ErrInvalidReceipt
	}
	return nil
}
