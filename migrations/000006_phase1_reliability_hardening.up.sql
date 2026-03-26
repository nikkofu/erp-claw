alter table outbox add column if not exists attempts integer not null default 0;
alter table outbox add column if not exists last_error text;
alter table outbox add column if not exists processing_at timestamptz;

create index if not exists idx_outbox_pending_available on outbox(status, available_at, id);
