\set ON_ERROR_STOP on

\connect paperstacks

-- Prerequisites (server-level, before this migration runs):
--   1. pg_cron must be installed on the PostgreSQL server.
--   2. postgresql.conf must contain pg_cron in shared_preload_libraries.
--   3. cron.database_name must be 'paperstacks'.
--   4. PostgreSQL must have been restarted after changing those settings.
--
-- Recommended for this project:
--   cron.use_background_workers = on
-- This lets pg_cron start jobs as PostgreSQL background workers instead of
-- opening a separate local libpq connection that would require pg_hba/.pgpass
-- configuration for the job user.

-- CREATE EXTENSION requires a sufficiently privileged role and must be done
-- before switching to app_owner/app_rw.
CREATE EXTENSION IF NOT EXISTS pg_cron;

BEGIN;

SET ROLE app_owner;
SET search_path TO public;

-- Permanently remove entities that have been soft-deleted for longer than the
-- supplied retention period. The default retention period is 30 days.
--
-- SECURITY DEFINER is intentional: the scheduled job is owned by app_rw, but
-- the cleanup itself runs with the table-owning app_owner privileges. The
-- fixed search_path prevents object-shadowing attacks against SECURITY DEFINER.
CREATE OR REPLACE PROCEDURE public.cleanup_soft_deleted_entries(
    p_retention interval DEFAULT interval '30 days'
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, pg_temp
AS $$
DECLARE
    v_cutoff timestamptz;
BEGIN
    IF p_retention IS NULL OR p_retention < interval '0 seconds' THEN
        RAISE EXCEPTION 'Retention period must be a non-negative interval';
    END IF;

    -- statement_timestamp() makes one stable cutoff for the complete run.
    v_cutoff := statement_timestamp() - p_retention;

    -- Lock cleanup candidates first. This prevents a concurrent restore from
    -- changing deleted_at while related rows are already being removed.
    PERFORM p.uuid
    FROM public.paper AS p
    WHERE p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff
    ORDER BY p.uuid
    FOR UPDATE;

    PERFORM s.uuid
    FROM public.stack AS s
    WHERE s.deleted_at IS NOT NULL
      AND s.deleted_at < v_cutoff
    ORDER BY s.uuid
    FOR UPDATE;

    -- paper_author currently uses ON DELETE RESTRICT, therefore its rows must
    -- be deleted explicitly before deleting the corresponding papers.
    DELETE FROM public.paper_author AS pa
    USING public.paper AS p
    WHERE pa.uuid_paper = p.uuid
      AND p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff;

    -- Existing foreign keys remove stack_paper, pdf and metadata rows through
    -- ON DELETE CASCADE when a paper is permanently deleted.
    DELETE FROM public.paper AS p
    WHERE p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff;

    -- Existing stack_paper FK removes stack relationships through
    -- ON DELETE CASCADE when a stack is permanently deleted.
    DELETE FROM public.stack AS s
    WHERE s.deleted_at IS NOT NULL
      AND s.deleted_at < v_cutoff;
END;
$$;

COMMENT ON PROCEDURE public.cleanup_soft_deleted_entries(interval) IS
'Permanently deletes papers and stacks whose soft-delete timestamp is older than the supplied retention period (default: 30 days).';

REVOKE ALL ON PROCEDURE public.cleanup_soft_deleted_entries(interval) FROM PUBLIC;
GRANT EXECUTE ON PROCEDURE public.cleanup_soft_deleted_entries(interval) TO app_rw;

RESET ROLE;

-- app_rw owns the scheduled job and only needs access to the pg_cron schema and
-- its scheduling function. The cleanup procedure itself runs as app_owner.
GRANT USAGE ON SCHEMA cron TO app_rw;
GRANT EXECUTE ON FUNCTION cron.schedule(text, text, text) TO app_rw;

-- Create the daily cleanup job as app_rw so the job does not run as a
-- superuser. With cron.timezone = 'UTC', "0 3 * * *" means 03:00 UTC daily.
SET ROLE app_rw;

SELECT cron.schedule(
    'soft-delete-cleanup',
    '0 3 * * *',
    $$CALL public.cleanup_soft_deleted_entries(interval '30 days');$$
);

RESET ROLE;

COMMIT;

-- Monitoring examples:
--   SELECT * FROM cron.job WHERE jobname = 'soft-delete-cleanup';
--   SELECT *
--   FROM cron.job_run_details
--   WHERE jobid IN (
--       SELECT jobid FROM cron.job WHERE jobname = 'soft-delete-cleanup'
--   )
--   ORDER BY start_time DESC
--   LIMIT 20;
