create table if not exists payable_payment_plan (
    id text primary key,
    tenant_id text not null,
    payable_bill_id text not null,
    status text not null default 'planned',
    due_date date not null,
    created_by text not null,
    created_at timestamptz not null default now(),
    unique (tenant_id, id),
    foreign key (tenant_id, payable_bill_id) references payable_bill(tenant_id, id) on delete cascade
);
