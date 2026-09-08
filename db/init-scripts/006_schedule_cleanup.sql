\set ON_ERROR_STOP on

\connect paperstacks

SELECT cron.schedule(
    'soft-delete-cleanup',
    '0 3 * * *',
    $$CALL public.cleanup_soft_deleted_entries(interval '30 days');$$
);