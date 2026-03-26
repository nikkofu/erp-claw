drop index if exists idx_agent_task_tenant_status_queued_at;
drop index if exists idx_agent_task_tenant_session_queued_at;

alter table agent_task
    drop constraint if exists agent_task_status_check;

alter table agent_session
    drop constraint if exists agent_session_status_check;
