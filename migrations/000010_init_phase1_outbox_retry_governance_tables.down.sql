drop index if exists idx_outbox_status_available_at;

alter table if exists outbox
    drop column if exists failed_at,
    drop column if exists last_error,
    drop column if exists attempts;
