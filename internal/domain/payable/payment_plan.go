package payable

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidPaymentPlan = errors.New("invalid payable payment plan")

type PaymentPlanStatus string

const (
	PaymentPlanStatusPlanned PaymentPlanStatus = "planned"
)

type PaymentPlan struct {
	ID             string
	TenantID       string
	PayableBillID  string
	Status         PaymentPlanStatus
	DueDateISO8601 string
	CreatedBy      string
}

func NewPaymentPlan(id, tenantID, payableBillID, createdBy, dueDateISO8601 string) (PaymentPlan, error) {
	plan := PaymentPlan{
		ID:             strings.TrimSpace(id),
		TenantID:       strings.TrimSpace(tenantID),
		PayableBillID:  strings.TrimSpace(payableBillID),
		Status:         PaymentPlanStatusPlanned,
		DueDateISO8601: strings.TrimSpace(dueDateISO8601),
		CreatedBy:      strings.TrimSpace(createdBy),
	}
	if plan.ID == "" || plan.TenantID == "" || plan.PayableBillID == "" || plan.DueDateISO8601 == "" || plan.CreatedBy == "" {
		return PaymentPlan{}, ErrInvalidPaymentPlan
	}
	dueDate, err := time.Parse("2006-01-02", plan.DueDateISO8601)
	if err != nil {
		return PaymentPlan{}, ErrInvalidPaymentPlan
	}
	plan.DueDateISO8601 = dueDate.Format("2006-01-02")
	return plan, nil
}
