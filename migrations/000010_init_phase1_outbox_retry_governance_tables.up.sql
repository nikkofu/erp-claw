alter table if exists outbox
    add column if not exists attempts int not null default 0,
    add column if not exists last_error text,
    add column if not exists failed_at timestamptz;

create index if not exists idx_outbox_status_available_at
    on outbox (status, available_at);
