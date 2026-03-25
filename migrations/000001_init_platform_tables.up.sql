create table if not exists tenant (
    id bigserial primary key,
    code text not null unique,
    name text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists tenant_cell (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    cell_key text not null,
    db_schema text not null,
    region text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, cell_key)
);

create table if not exists audit_log (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    request_id text not null,
    actor_id text not null,
    action text not null,
    resource text not null,
    payload jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create table if not exists agent_session (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    session_key text not null,
    status text not null default 'open',
    metadata jsonb not null default '{}'::jsonb,
    started_at timestamptz not null default now(),
    ended_at timestamptz,
    unique (tenant_id, session_key)
);

create table if not exists agent_task (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    session_id bigint references agent_session(id) on delete set null,
    task_type text not null,
    status text not null default 'pending',
    input jsonb not null default '{}'::jsonb,
    output jsonb not null default '{}'::jsonb,
    attempts int not null default 0,
    queued_at timestamptz not null default now(),
    completed_at timestamptz
);

create table if not exists outbox (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    topic text not null,
    event_type text not null,
    payload jsonb not null,
    status text not null default 'pending',
    available_at timestamptz not null default now(),
    published_at timestamptz,
    created_at timestamptz not null default now()
);
