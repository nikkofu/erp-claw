create table if not exists inbox (
    id bigserial primary key,
    tenant_id bigint not null references tenant(id) on delete cascade,
    message_key text not null,
    topic text not null,
    payload jsonb not null default '{}'::jsonb,
    status text not null default 'received',
    error text,
    received_at timestamptz not null default now(),
    processed_at timestamptz,
    created_at timestamptz not null default now(),
    unique (tenant_id, message_key)
);

create index if not exists idx_inbox_status_received_at
    on inbox (status, received_at);
