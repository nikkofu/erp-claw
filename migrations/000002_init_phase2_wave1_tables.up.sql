create table if not exists supplier (
    id text primary key,
    tenant_id text not null,
    code text not null,
    name text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code)
);

create table if not exists product (
    id text primary key,
    tenant_id text not null,
    sku text not null,
    name text not null,
    unit text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, sku)
);

create table if not exists warehouse (
    id text primary key,
    tenant_id text not null,
    code text not null,
    name text not null,
    status text not null default 'active',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code)
);

create table if not exists approval_request (
    id text primary key,
    tenant_id text not null,
    resource_type text not null,
    resource_id text not null,
    status text not null default 'pending',
    requested_by text not null,
    decided_by text,
    created_at timestamptz not null default now(),
    decided_at timestamptz,
    unique (tenant_id, resource_type, resource_id)
);

create table if not exists purchase_order (
    id text primary key,
    tenant_id text not null,
    supplier_id text not null references supplier(id),
    warehouse_id text not null references warehouse(id),
    approval_id text references approval_request(id),
    status text not null default 'draft',
    created_by text not null default '',
    submitted_by text,
    approved_by text,
    rejected_by text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists purchase_order_line (
    id bigserial primary key,
    purchase_order_id text not null references purchase_order(id) on delete cascade,
    tenant_id text not null,
    line_no int not null,
    product_id text not null references product(id),
    quantity int not null,
    created_at timestamptz not null default now(),
    unique (purchase_order_id, line_no)
);
