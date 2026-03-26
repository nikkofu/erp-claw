create table if not exists agent_profile_allowed_model (
    tenant_id text not null,
    agent_profile_id text not null,
    model_entry_id text not null,
    created_at timestamptz not null default now(),
    primary key (tenant_id, agent_profile_id, model_entry_id),
    foreign key (tenant_id, agent_profile_id) references agent_profile(tenant_id, id) on delete cascade,
    foreign key (tenant_id, model_entry_id) references model_catalog_entries(tenant_id, entry_id) on delete cascade
);

create table if not exists agent_profile_allowed_tool (
    tenant_id text not null,
    agent_profile_id text not null,
    tool_entry_id text not null,
    created_at timestamptz not null default now(),
    primary key (tenant_id, agent_profile_id, tool_entry_id),
    foreign key (tenant_id, agent_profile_id) references agent_profile(tenant_id, id) on delete cascade,
    foreign key (tenant_id, tool_entry_id) references tool_catalog_entries(tenant_id, entry_id) on delete cascade
);
