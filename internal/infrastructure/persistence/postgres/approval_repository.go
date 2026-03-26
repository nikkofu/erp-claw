package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	domainapproval "github.com/nikkofu/erp-claw/internal/domain/approval"
)

var errApprovalRepositoryNilDB = errors.New("postgres approval repository requires non-nil db")

type ApprovalRepository struct {
	db *sql.DB
}

func NewApprovalRepository(db *sql.DB) (*ApprovalRepository, error) {
	if db == nil {
		return nil, errApprovalRepositoryNilDB
	}

	return &ApprovalRepository{db: db}, nil
}

func (r *ApprovalRepository) SaveDefinition(ctx context.Context, definition domainapproval.Definition) (domainapproval.Definition, error) {
	if strings.TrimSpace(definition.ID) == "" {
		definition.ID = uuid.NewString()
	}

	if err := r.db.QueryRowContext(
		ctx,
		`insert into approval_definition(tenant_id, id, name, approver_id, active)
		 values ($1, $2, $3, $4, $5)
		 on conflict (tenant_id, id) do update
		 set name = excluded.name,
		     approver_id = excluded.approver_id,
		     active = excluded.active,
		     updated_at = now()
		 returning id`,
		definition.TenantID,
		definition.ID,
		definition.Name,
		definition.ApproverID,
		definition.Active,
	).Scan(&definition.ID); err != nil {
		return domainapproval.Definition{}, err
	}

	return definition, nil
}

func (r *ApprovalRepository) ListDefinitions(ctx context.Context, tenantID string) ([]domainapproval.Definition, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, name, approver_id, active, created_at, updated_at
		 from approval_definition
		 where tenant_id = $1
		 order by created_at asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	definitions := make([]domainapproval.Definition, 0)
	for rows.Next() {
		definition := domainapproval.Definition{TenantID: tenantID}
		if err := rows.Scan(&definition.ID, &definition.Name, &definition.ApproverID, &definition.Active, &definition.CreatedAt, &definition.UpdatedAt); err != nil {
			return nil, err
		}
		definitions = append(definitions, definition)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return definitions, nil
}

func (r *ApprovalRepository) GetDefinitionByID(ctx context.Context, tenantID, definitionID string) (domainapproval.Definition, error) {
	definition := domainapproval.Definition{TenantID: tenantID}
	if err := r.db.QueryRowContext(
		ctx,
		`select id, name, approver_id, active, created_at, updated_at
		 from approval_definition
		 where tenant_id = $1 and id = $2`,
		tenantID,
		definitionID,
	).Scan(&definition.ID, &definition.Name, &definition.ApproverID, &definition.Active, &definition.CreatedAt, &definition.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainapproval.Definition{}, domainapproval.ErrDefinitionNotFound
		}
		return domainapproval.Definition{}, err
	}

	return definition, nil
}

func (r *ApprovalRepository) CreateInstance(ctx context.Context, instance domainapproval.Instance) (domainapproval.Instance, error) {
	if strings.TrimSpace(instance.ID) == "" {
		instance.ID = uuid.NewString()
	}

	var decidedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`insert into approval_instance(tenant_id, id, definition_id, resource_type, resource_id, requested_by, status)
		 values ($1, $2, $3, $4, $5, $6, $7)
		 returning id, created_at, decided_at`,
		instance.TenantID,
		instance.ID,
		instance.DefinitionID,
		instance.ResourceType,
		instance.ResourceID,
		instance.RequestedBy,
		string(instance.Status),
	).Scan(&instance.ID, &instance.CreatedAt, &decidedAt); err != nil {
		return domainapproval.Instance{}, err
	}
	if decidedAt.Valid {
		instance.DecidedAt = &decidedAt.Time
	}

	return instance, nil
}

func (r *ApprovalRepository) ListInstances(ctx context.Context, tenantID string) ([]domainapproval.Instance, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, definition_id, resource_type, resource_id, requested_by, status, created_at, decided_at
		 from approval_instance
		 where tenant_id = $1
		 order by created_at asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]domainapproval.Instance, 0)
	for rows.Next() {
		instance := domainapproval.Instance{TenantID: tenantID}
		var status string
		var decidedAt sql.NullTime
		if err := rows.Scan(&instance.ID, &instance.DefinitionID, &instance.ResourceType, &instance.ResourceID, &instance.RequestedBy, &status, &instance.CreatedAt, &decidedAt); err != nil {
			return nil, err
		}
		instance.Status = domainapproval.InstanceStatus(status)
		if decidedAt.Valid {
			instance.DecidedAt = &decidedAt.Time
		}
		instances = append(instances, instance)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *ApprovalRepository) GetInstanceByID(ctx context.Context, tenantID, instanceID string) (domainapproval.Instance, error) {
	instance := domainapproval.Instance{TenantID: tenantID}
	var status string
	var decidedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`select id, definition_id, resource_type, resource_id, requested_by, status, created_at, decided_at
		 from approval_instance
		 where tenant_id = $1 and id = $2`,
		tenantID,
		instanceID,
	).Scan(&instance.ID, &instance.DefinitionID, &instance.ResourceType, &instance.ResourceID, &instance.RequestedBy, &status, &instance.CreatedAt, &decidedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainapproval.Instance{}, domainapproval.ErrInstanceNotFound
		}
		return domainapproval.Instance{}, err
	}
	instance.Status = domainapproval.InstanceStatus(status)
	if decidedAt.Valid {
		instance.DecidedAt = &decidedAt.Time
	}

	return instance, nil
}

func (r *ApprovalRepository) UpdateInstanceStatus(ctx context.Context, tenantID, instanceID string, status domainapproval.InstanceStatus) error {
	result, err := r.db.ExecContext(
		ctx,
		`update approval_instance
		 set status = $3, decided_at = now()
		 where tenant_id = $1 and id = $2`,
		tenantID,
		instanceID,
		string(status),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domainapproval.ErrInstanceNotFound
	}
	return nil
}

func (r *ApprovalRepository) DeleteInstance(ctx context.Context, tenantID, instanceID string) error {
	result, err := r.db.ExecContext(
		ctx,
		`delete from approval_instance
		 where tenant_id = $1 and id = $2`,
		tenantID,
		instanceID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domainapproval.ErrInstanceNotFound
	}
	return nil
}

func (r *ApprovalRepository) CreateTask(ctx context.Context, task domainapproval.Task) (domainapproval.Task, error) {
	if strings.TrimSpace(task.ID) == "" {
		task.ID = uuid.NewString()
	}

	var decidedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`insert into approval_task(tenant_id, id, instance_id, approver_id, status, decided_by, comment)
		 values ($1, $2, $3, $4, $5, nullif($6, ''), nullif($7, ''))
		 returning id, created_at, decided_at`,
		task.TenantID,
		task.ID,
		task.InstanceID,
		task.ApproverID,
		string(task.Status),
		task.DecidedBy,
		task.Comment,
	).Scan(&task.ID, &task.CreatedAt, &decidedAt); err != nil {
		return domainapproval.Task{}, err
	}
	if decidedAt.Valid {
		task.DecidedAt = &decidedAt.Time
	}

	return task, nil
}

func (r *ApprovalRepository) ListTasks(ctx context.Context, tenantID string) ([]domainapproval.Task, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, instance_id, approver_id, status, coalesce(decided_by, ''), coalesce(comment, ''), created_at, decided_at
		 from approval_task
		 where tenant_id = $1
		 order by created_at asc, id asc`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domainapproval.Task, 0)
	for rows.Next() {
		task := domainapproval.Task{TenantID: tenantID}
		var status string
		var decidedAt sql.NullTime
		if err := rows.Scan(&task.ID, &task.InstanceID, &task.ApproverID, &status, &task.DecidedBy, &task.Comment, &task.CreatedAt, &decidedAt); err != nil {
			return nil, err
		}
		task.Status = domainapproval.TaskStatus(status)
		if decidedAt.Valid {
			task.DecidedAt = &decidedAt.Time
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *ApprovalRepository) GetTaskByID(ctx context.Context, tenantID, taskID string) (domainapproval.Task, error) {
	task := domainapproval.Task{TenantID: tenantID}
	var status string
	var decidedAt sql.NullTime
	if err := r.db.QueryRowContext(
		ctx,
		`select id, instance_id, approver_id, status, coalesce(decided_by, ''), coalesce(comment, ''), created_at, decided_at
		 from approval_task
		 where tenant_id = $1 and id = $2`,
		tenantID,
		taskID,
	).Scan(&task.ID, &task.InstanceID, &task.ApproverID, &status, &task.DecidedBy, &task.Comment, &task.CreatedAt, &decidedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainapproval.Task{}, domainapproval.ErrTaskNotFound
		}
		return domainapproval.Task{}, err
	}
	task.Status = domainapproval.TaskStatus(status)
	if decidedAt.Valid {
		task.DecidedAt = &decidedAt.Time
	}

	return task, nil
}

func (r *ApprovalRepository) UpdateTaskDecision(ctx context.Context, tenantID, taskID string, status domainapproval.TaskStatus, decidedBy, comment string) error {
	result, err := r.db.ExecContext(
		ctx,
		`update approval_task
		 set status = $3,
		     decided_by = nullif($4, ''),
		     comment = nullif($5, ''),
		     decided_at = now()
		 where tenant_id = $1 and id = $2`,
		tenantID,
		taskID,
		string(status),
		decidedBy,
		comment,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domainapproval.ErrTaskNotFound
	}
	return nil
}

func (r *ApprovalRepository) ListTasksByInstance(ctx context.Context, tenantID, instanceID string) ([]domainapproval.Task, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`select id, approver_id, status, coalesce(decided_by, ''), coalesce(comment, ''), created_at, decided_at
		 from approval_task
		 where tenant_id = $1 and instance_id = $2
		 order by created_at asc, id asc`,
		tenantID,
		instanceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domainapproval.Task, 0)
	for rows.Next() {
		task := domainapproval.Task{TenantID: tenantID, InstanceID: instanceID}
		var status string
		var decidedAt sql.NullTime
		if err := rows.Scan(&task.ID, &task.ApproverID, &status, &task.DecidedBy, &task.Comment, &task.CreatedAt, &decidedAt); err != nil {
			return nil, err
		}
		task.Status = domainapproval.TaskStatus(status)
		if decidedAt.Valid {
			task.DecidedAt = &decidedAt.Time
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
