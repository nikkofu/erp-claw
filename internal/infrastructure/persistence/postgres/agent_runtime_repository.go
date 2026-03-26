package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	agentruntime "github.com/nikkofu/erp-claw/internal/domain/agentruntime"
)

var errAgentRuntimeRepositoryNilDB = errors.New("postgres agent runtime repository requires non-nil db")

type AgentRuntimeRepository struct {
	db *sql.DB
}

func NewAgentRuntimeRepository(db *sql.DB) (*AgentRuntimeRepository, error) {
	if db == nil {
		return nil, errAgentRuntimeRepositoryNilDB
	}

	return &AgentRuntimeRepository{db: db}, nil
}

func (r *AgentRuntimeRepository) CreateSession(ctx context.Context, session agentruntime.Session) (agentruntime.Session, error) {
	tenantID, err := parseInt64ID(session.TenantID)
	if err != nil {
		return agentruntime.Session{}, err
	}

	metadata, err := json.Marshal(session.Metadata)
	if err != nil {
		return agentruntime.Session{}, err
	}

	var dbID int64
	var startedAt time.Time
	var endedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`insert into agent_session(tenant_id, session_key, status, metadata)
		 values ($1, $2, $3, $4)
		 returning id, started_at, ended_at`,
		tenantID,
		session.SessionKey,
		string(session.Status),
		metadata,
	).Scan(&dbID, &startedAt, &endedAt); err != nil {
		return agentruntime.Session{}, err
	}

	session.ID = strconv.FormatInt(dbID, 10)
	session.StartedAt = startedAt
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	return session, nil
}

func (r *AgentRuntimeRepository) GetSessionByTenantAndKey(ctx context.Context, tenantID, sessionKey string) (agentruntime.Session, error) {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return agentruntime.Session{}, err
	}

	var dbID int64
	var status string
	var metadataRaw []byte
	var startedAt time.Time
	var endedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`select id, status, metadata, started_at, ended_at
		 from agent_session
		 where tenant_id = $1 and session_key = $2`,
		parsedTenantID,
		sessionKey,
	).Scan(&dbID, &status, &metadataRaw, &startedAt, &endedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return agentruntime.Session{}, agentruntime.ErrSessionNotFound
		}
		return agentruntime.Session{}, err
	}

	metadata := map[string]any{}
	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
			return agentruntime.Session{}, err
		}
	}

	session := agentruntime.Session{
		ID:         strconv.FormatInt(dbID, 10),
		TenantID:   tenantID,
		SessionKey: sessionKey,
		Status:     agentruntime.SessionStatus(status),
		Metadata:   metadata,
		StartedAt:  startedAt,
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	return session, nil
}

func (r *AgentRuntimeRepository) ListSessions(ctx context.Context, tenantID string) ([]agentruntime.Session, error) {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`select id, session_key, status, metadata, started_at, ended_at
		 from agent_session
		 where tenant_id = $1
		 order by started_at desc, id desc`,
		parsedTenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]agentruntime.Session, 0)
	for rows.Next() {
		var dbID int64
		var sessionKey string
		var status string
		var metadataRaw []byte
		var startedAt time.Time
		var endedAt sql.NullTime
		if err := rows.Scan(&dbID, &sessionKey, &status, &metadataRaw, &startedAt, &endedAt); err != nil {
			return nil, err
		}

		metadata := map[string]any{}
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &metadata); err != nil {
				return nil, err
			}
		}

		session := agentruntime.Session{
			ID:         strconv.FormatInt(dbID, 10),
			TenantID:   tenantID,
			SessionKey: sessionKey,
			Status:     agentruntime.SessionStatus(status),
			Metadata:   metadata,
			StartedAt:  startedAt,
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (r *AgentRuntimeRepository) UpdateSessionStatus(ctx context.Context, tenantID, sessionKey string, status agentruntime.SessionStatus, endedAt *time.Time) error {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(
		ctx,
		`update agent_session
		 set status = $3, ended_at = $4
		 where tenant_id = $1 and session_key = $2`,
		parsedTenantID,
		sessionKey,
		string(status),
		endedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return agentruntime.ErrSessionNotFound
	}

	return nil
}

func (r *AgentRuntimeRepository) CreateTask(ctx context.Context, task agentruntime.Task) (agentruntime.Task, error) {
	tenantID, err := parseInt64ID(task.TenantID)
	if err != nil {
		return agentruntime.Task{}, err
	}

	input, err := json.Marshal(task.Input)
	if err != nil {
		return agentruntime.Task{}, err
	}
	output, err := json.Marshal(task.Output)
	if err != nil {
		return agentruntime.Task{}, err
	}

	var sessionID any
	if strings.TrimSpace(task.SessionID) != "" {
		parsedSessionID, err := r.lookupSessionDBID(ctx, tenantID, task.SessionID)
		if err != nil {
			return agentruntime.Task{}, err
		}
		sessionID = parsedSessionID
	}

	var dbID int64
	var queuedAt time.Time
	var completedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`insert into agent_task(tenant_id, session_id, task_type, status, input, output, attempts)
		 values ($1, $2, $3, $4, $5, $6, $7)
		 returning id, queued_at, completed_at`,
		tenantID,
		sessionID,
		task.TaskType,
		string(task.Status),
		input,
		output,
		task.Attempts,
	).Scan(&dbID, &queuedAt, &completedAt); err != nil {
		return agentruntime.Task{}, err
	}

	task.ID = strconv.FormatInt(dbID, 10)
	task.QueuedAt = queuedAt
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	return task, nil
}

func (r *AgentRuntimeRepository) GetTaskByID(ctx context.Context, tenantID, taskID string) (agentruntime.Task, error) {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return agentruntime.Task{}, err
	}
	parsedTaskID, err := parseInt64ID(taskID)
	if err != nil {
		return agentruntime.Task{}, err
	}

	var sessionID sql.NullInt64
	var sessionKey sql.NullString
	var taskType string
	var status string
	var inputRaw []byte
	var outputRaw []byte
	var attempts int
	var queuedAt time.Time
	var completedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`select t.session_id,
		        coalesce(s.session_key, ''),
		        t.task_type,
		        t.status,
		        t.input,
		        t.output,
		        t.attempts,
		        t.queued_at,
		        t.completed_at
		 from agent_task t
		 left join agent_session s
		   on s.tenant_id = t.tenant_id
		  and s.id = t.session_id
		 where t.tenant_id = $1 and t.id = $2`,
		parsedTenantID,
		parsedTaskID,
	).Scan(&sessionID, &sessionKey, &taskType, &status, &inputRaw, &outputRaw, &attempts, &queuedAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return agentruntime.Task{}, agentruntime.ErrTaskNotFound
		}
		return agentruntime.Task{}, err
	}

	input := map[string]any{}
	if len(inputRaw) > 0 {
		if err := json.Unmarshal(inputRaw, &input); err != nil {
			return agentruntime.Task{}, err
		}
	}
	output := map[string]any{}
	if len(outputRaw) > 0 {
		if err := json.Unmarshal(outputRaw, &output); err != nil {
			return agentruntime.Task{}, err
		}
	}

	task := agentruntime.Task{
		ID:       taskID,
		TenantID: tenantID,
		TaskType: taskType,
		Status:   agentruntime.TaskStatus(status),
		Input:    input,
		Output:   output,
		Attempts: attempts,
		QueuedAt: queuedAt,
	}
	if sessionID.Valid {
		task.SessionID = sessionKey.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

func (r *AgentRuntimeRepository) ListTasks(ctx context.Context, tenantID, sessionID string) ([]agentruntime.Task, error) {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return nil, err
	}

	query := `select t.id,
	                 coalesce(s.session_key, ''),
	                 t.task_type,
	                 t.status,
	                 t.input,
	                 t.output,
	                 t.attempts,
	                 t.queued_at,
	                 t.completed_at
		from agent_task t
		left join agent_session s
		  on s.tenant_id = t.tenant_id
		 and s.id = t.session_id
		where t.tenant_id = $1`
	args := []any{parsedTenantID}
	if strings.TrimSpace(sessionID) != "" {
		query += ` and s.session_key = $2`
		args = append(args, strings.TrimSpace(sessionID))
	}
	query += ` order by t.queued_at desc, t.id desc`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]agentruntime.Task, 0)
	for rows.Next() {
		var dbID int64
		var sessionKey sql.NullString
		var taskType string
		var status string
		var inputRaw []byte
		var outputRaw []byte
		var attempts int
		var queuedAt time.Time
		var completedAt sql.NullTime
		if err := rows.Scan(&dbID, &sessionKey, &taskType, &status, &inputRaw, &outputRaw, &attempts, &queuedAt, &completedAt); err != nil {
			return nil, err
		}

		input := map[string]any{}
		if len(inputRaw) > 0 {
			if err := json.Unmarshal(inputRaw, &input); err != nil {
				return nil, err
			}
		}
		output := map[string]any{}
		if len(outputRaw) > 0 {
			if err := json.Unmarshal(outputRaw, &output); err != nil {
				return nil, err
			}
		}

		task := agentruntime.Task{
			ID:       strconv.FormatInt(dbID, 10),
			TenantID: tenantID,
			TaskType: taskType,
			Status:   agentruntime.TaskStatus(status),
			Input:    input,
			Output:   output,
			Attempts: attempts,
			QueuedAt: queuedAt,
		}
		if sessionKey.Valid {
			task.SessionID = sessionKey.String
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *AgentRuntimeRepository) UpdateTaskStatus(ctx context.Context, tenantID, taskID string, status agentruntime.TaskStatus, output map[string]any, completedAt *time.Time) error {
	parsedTenantID, err := parseInt64ID(tenantID)
	if err != nil {
		return err
	}
	parsedTaskID, err := parseInt64ID(taskID)
	if err != nil {
		return err
	}

	outputRaw, err := json.Marshal(output)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(
		ctx,
		`update agent_task
		 set status = $3,
		     output = $4,
		     completed_at = $5,
		     attempts = case when $3 = 'running' then attempts + 1 else attempts end
		 where tenant_id = $1 and id = $2`,
		parsedTenantID,
		parsedTaskID,
		string(status),
		outputRaw,
		completedAt,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return agentruntime.ErrTaskNotFound
	}

	return nil
}

func parseInt64ID(raw string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
}

func (r *AgentRuntimeRepository) lookupSessionDBID(ctx context.Context, tenantID int64, sessionKey string) (int64, error) {
	var sessionID int64
	if err := r.db.QueryRowContext(
		ctx,
		`select id
		 from agent_session
		 where tenant_id = $1 and session_key = $2`,
		tenantID,
		strings.TrimSpace(sessionKey),
	).Scan(&sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, agentruntime.ErrSessionNotFound
		}
		return 0, err
	}

	return sessionID, nil
}
