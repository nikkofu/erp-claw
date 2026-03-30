package runtime

import "time"

type TaskListQuery struct {
	TenantID  string
	ActorID   string
	SessionID string
	Status    TaskStatus
	Limit     int
	Cursor    string
}

type SessionListQuery struct {
	TenantID string
	ActorID  string
	Status   SessionStatus
	Limit    int
}

type TaskListPage struct {
	Items      []Task
	NextCursor string
	AsOf       time.Time
}

type SessionListPage struct {
	Items []Session
	AsOf  time.Time
}
