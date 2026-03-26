create table if not exists approval_definition (
    tenant_id text not null,
    id text not null,
    name text not null,
    approver_id text not null,
    active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id)
);

create table if not exists approval_instance (
    tenant_id text not null,
    id text not null,
    definition_id text not null,
    resource_type text not null,
    resource_id text not null,
    requested_by text not null,
    status text not null,
    created_at timestamptz not null default now(),
    decided_at timestamptz,
    primary key (tenant_id, id),
    foreign key (tenant_id, definition_id) references approval_definition(tenant_id, id) on delete cascade
);

create table if not exists approval_task (
    tenant_id text not null,
    id text not null,
    instance_id text not null,
    approver_id text not null,
    status text not null,
    decided_by text,
    comment text,
    created_at timestamptz not null default now(),
    decided_at timestamptz,
    primary key (tenant_id, id),
    foreign key (tenant_id, instance_id) references approval_instance(tenant_id, id) on delete cascade
);
