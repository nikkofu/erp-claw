create table if not exists tool_catalog_entries (
    tenant_id text not null,
    entry_id text not null,
    tool_key text not null,
    display_name text not null,
    risk_level text not null,
    status text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, entry_id)
);
