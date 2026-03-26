create table if not exists transfer_order (
    id text primary key,
    tenant_id text not null,
    product_id text not null,
    from_warehouse_id text not null,
    to_warehouse_id text not null,
    quantity int not null,
    status text not null default 'planned',
    created_by text not null,
    executed_by text,
    executed_at timestamptz,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, product_id) references product(tenant_id, id),
    foreign key (tenant_id, from_warehouse_id) references warehouse(tenant_id, id),
    foreign key (tenant_id, to_warehouse_id) references warehouse(tenant_id, id)
);
