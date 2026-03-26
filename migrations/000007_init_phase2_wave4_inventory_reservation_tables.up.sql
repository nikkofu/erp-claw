create table if not exists inventory_reservation (
    id text primary key,
    tenant_id text not null,
    product_id text not null,
    warehouse_id text not null,
    reference_type text not null,
    reference_id text not null,
    status text not null default 'active',
    created_by text not null,
    quantity int not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, product_id) references product(tenant_id, id),
    foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)
);
