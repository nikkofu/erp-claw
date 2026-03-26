create table if not exists iam_department (
    tenant_id text not null,
    id text not null,
    name text not null,
    parent_department_id text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (tenant_id, id),
    foreign key (tenant_id, parent_department_id) references iam_department(tenant_id, id) on delete set null
);

create table if not exists iam_user_role (
    tenant_id text not null,
    id text not null,
    user_id text not null,
    role_id text not null,
    assigned_at timestamptz not null default now(),
    primary key (tenant_id, id),
    unique (tenant_id, user_id, role_id),
    foreign key (tenant_id, user_id) references iam_user(tenant_id, id) on delete cascade,
    foreign key (tenant_id, role_id) references iam_role(tenant_id, id) on delete cascade
);

insert into iam_user_role (tenant_id, id, user_id, role_id, assigned_at)
select tenant_id, id, user_id, role_id, bound_at
from iam_user_role_binding
on conflict (tenant_id, user_id, role_id) do nothing;

create table if not exists iam_user_department (
    tenant_id text not null,
    id text not null,
    user_id text not null,
    department_id text not null,
    assigned_at timestamptz not null default now(),
    primary key (tenant_id, id),
    unique (tenant_id, user_id, department_id),
    foreign key (tenant_id, user_id) references iam_user(tenant_id, id) on delete cascade,
    foreign key (tenant_id, department_id) references iam_department(tenant_id, id) on delete cascade
);
