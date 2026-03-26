create table if not exists sales_order (
    id text primary key,
    tenant_id text not null,
    warehouse_id text not null,
    external_ref text not null,
    status text not null default 'draft',
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    unique (tenant_id, external_ref),
    foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)
);

create table if not exists sales_order_line (
    id bigserial primary key,
    tenant_id text not null,
    sales_order_id text not null,
    line_no int not null,
    product_id text not null,
    quantity int not null,
    created_at timestamptz not null default now(),
    unique (sales_order_id, line_no),
    foreign key (tenant_id, sales_order_id) references sales_order(tenant_id, id) on delete cascade,
    foreign key (tenant_id, product_id) references product(tenant_id, id)
);
