do $$
begin
    if not exists (
        select 1
        from pg_constraint
        where conname = 'agent_session_status_check'
    ) then
        alter table agent_session
            add constraint agent_session_status_check
            check (status in ('open', 'closed'));
    end if;
end
$$;

do $$
begin
    if not exists (
        select 1
        from pg_constraint
        where conname = 'agent_task_status_check'
    ) then
        alter table agent_task
            add constraint agent_task_status_check
            check (status in ('pending', 'running', 'succeeded', 'failed', 'canceled'));
    end if;
end
$$;

create index if not exists idx_agent_task_tenant_session_queued_at
    on agent_task (tenant_id, session_id, queued_at desc);

create index if not exists idx_agent_task_tenant_status_queued_at
    on agent_task (tenant_id, status, queued_at desc);
