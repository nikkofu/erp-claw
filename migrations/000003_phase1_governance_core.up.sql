create table if not exists policy_rule (
    tenant_id text not null,
    id text not null,
    command_name text not null,
    actor_id text not null default '*',
    decision text not null,
    priority int not null default 100,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id),
    check (decision in ('ALLOW', 'ALLOW_WITH_GUARD', 'REQUIRE_APPROVAL', 'DENY'))
);

create index if not exists idx_policy_rule_lookup
    on policy_rule (tenant_id, command_name, actor_id, is_active, priority desc);

create table if not exists audit_event (
    tenant_id text not null,
    id text not null,
    command_name text not null,
    actor_id text not null,
    decision text not null,
    outcome text not null,
    error text not null default '',
    occurred_at timestamptz not null,
    recorded_at timestamptz not null default now(),
    primary key (tenant_id, id),
    check (decision in ('ALLOW', 'ALLOW_WITH_GUARD', 'REQUIRE_APPROVAL', 'DENY'))
);

create index if not exists idx_audit_event_lookup
    on audit_event (tenant_id, command_name, actor_id, occurred_at desc);
