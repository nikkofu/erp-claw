create table if not exists organization (
    tenant_id text not null,
    id text not null,
    name text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id)
);

create table if not exists iam_user (
    tenant_id text not null,
    id text not null,
    email text not null,
    display_name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id),
    unique (tenant_id, email)
);

create table if not exists iam_role (
    tenant_id text not null,
    id text not null,
    name text not null,
    description text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id)
);

create table if not exists iam_user_role_binding (
    tenant_id text not null,
    id text not null,
    user_id text not null,
    role_id text not null,
    bound_at timestamptz not null default now(),
    primary key (tenant_id, id),
    unique (tenant_id, user_id, role_id),
    foreign key (tenant_id, user_id) references iam_user(tenant_id, id) on delete cascade,
    foreign key (tenant_id, role_id) references iam_role(tenant_id, id) on delete cascade
);

create table if not exists agent_profile (
    tenant_id text not null,
    id text not null,
    name text not null,
    model text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id)
);
