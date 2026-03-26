package supplychain

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/nikkofu/erp-claw/internal/application/shared"
	"github.com/nikkofu/erp-claw/internal/domain/approval"
	"github.com/nikkofu/erp-claw/internal/domain/masterdata"
	"github.com/nikkofu/erp-claw/internal/domain/procurement"
)

type ServiceDeps struct {
	MasterData     masterdata.Repository
	PurchaseOrders procurement.Repository
	Approvals      approval.Repository
	Pipeline       *shared.Pipeline
}

type Service struct {
	masterData     masterdata.Repository
	purchaseOrders procurement.Repository
	approvals      approval.Repository
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
