create table if not exists payable_bill (
    id text primary key,
    tenant_id text not null,
    purchase_order_id text not null,
    status text not null default 'open',
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    unique (tenant_id, purchase_order_id),
    foreign key (tenant_id, purchase_order_id) references purchase_order(tenant_id, id)
);
