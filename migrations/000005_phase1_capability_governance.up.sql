create table if not exists model_catalog_entries (
    tenant_id text not null,
    entry_id text not null,
    model_key text not null,
    display_name text not null,
    provider text not null,
    status text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, entry_id)
);
