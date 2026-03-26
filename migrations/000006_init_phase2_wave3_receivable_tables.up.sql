create table if not exists receivable_bill (
    id text primary key,
    tenant_id text not null,
    external_ref text not null,
    status text not null default 'open',
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    unique (tenant_id, external_ref)
);
