create table if not exists receipt (
    id text primary key,
    tenant_id text not null,
    purchase_order_id text not null,
    warehouse_id text not null,
    status text not null default 'posted',
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id),
    foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id)
);

create table if not exists receipt_line (
    id bigserial primary key,
    tenant_id text not null,
    receipt_id text not null,
    line_no int not null,
    product_id text not null,
    quantity int not null,
    created_at timestamptz not null default now(),
    unique (receipt_id, line_no),
    foreign key (tenant_id, receipt_id) references receipt(tenant_id, id) on delete cascade,
    foreign key (tenant_id, product_id) references product(tenant_id, id)
);

create table if not exists inventory_ledger (
    id text primary key,
    tenant_id text not null,
    product_id text not null,
    warehouse_id text not null,
    movement_type text not null,
    quantity_delta int not null,
    reference_type text not null,
    reference_id text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, product_id) references product(tenant_id, id),
    foreign key (tenant_id, warehouse_id) references warehouse(tenant_id, id),
    foreign key (tenant_id, reference_id) references receipt(tenant_id, id)
);
