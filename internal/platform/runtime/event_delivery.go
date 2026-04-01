package runtime

import "time"

type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusRecovered DeliveryStatus = "recovered"
)

type DeliveryRecord struct {
	EventType    string
	TenantID     string
	SessionID    string
	TaskID       string
	AttemptCount int
	LastError    string
	Status       DeliveryStatus
	UpdatedAt    time.Time
}

type DeliveryListQuery struct {
	TenantID  string
	ActorID   string
	Status    DeliveryStatus
	SessionID string
	TaskID    string
	Limit     int
}

type DeliveryListPage struct {
	Items []DeliveryRecord
	AsOf  time.Time
}
