package approval

import (
	"errors"
	"strings"
)

var (
	ErrInvalidRequest     = errors.New("invalid approval request")
	ErrApprovalNotPending = errors.New("approval request is not pending")
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Request struct {
	ID           string
	TenantID     string
	ResourceType string
	ResourceID   string
	Status       Status
	RequestedBy  string
	DecidedBy    string
}

func NewRequest(id, tenantID, resourceType, resourceID, requestedBy string) (Request, error) {
	request := Request{
		ID:           strings.TrimSpace(id),
		TenantID:     strings.TrimSpace(tenantID),
		ResourceType: strings.TrimSpace(resourceType),
		ResourceID:   strings.TrimSpace(resourceID),
		Status:       StatusPending,
		RequestedBy:  strings.TrimSpace(requestedBy),
	}
	if request.ID == "" || request.TenantID == "" || request.ResourceType == "" || request.ResourceID == "" || request.RequestedBy == "" {
		return Request{}, ErrInvalidRequest
	}
	return request, nil
}

func (r *Request) Approve(actorID string) error {
	if r.Status != StatusPending {
		return ErrApprovalNotPending
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return ErrInvalidRequest
	}
	r.Status = StatusApproved
	r.DecidedBy = actorID
	return nil
}

func (r *Request) Reject(actorID string) error {
	if r.Status != StatusPending {
		return ErrApprovalNotPending
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return ErrInvalidRequest
	}
	r.Status = StatusRejected
	r.DecidedBy = actorID
	return nil
}
