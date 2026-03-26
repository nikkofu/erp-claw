drop index if exists idx_outbox_pending_available;

alter table outbox drop column if exists processing_at;
alter table outbox drop column if exists last_error;
alter table outbox drop column if exists attempts;
