\set ON_ERROR_STOP on

\connect paperstacks

BEGIN;

SET ROLE app_owner;
SET search_path TO public;

CREATE OR REPLACE PROCEDURE public.cleanup_soft_deleted_entries()
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public, pg_temp
AS $$
DECLARE
    retention CONSTANT interval := interval '30 days';
    v_cutoff timestamptz := statement_timestamp() - retention;
BEGIN


    -- paper_author uses ON DELETE RESTRICT.
    DELETE FROM public.paper_author AS pa
    USING public.paper AS p
    WHERE pa.uuid_paper = p.uuid
      AND p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff;

    -- stack_paper and metadata are cleaned via ON DELETE CASCADE.
    -- pdf_storage_key is stored directly on paper and is removed with the row.
    DELETE FROM public.paper AS p
    WHERE p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff;

    -- stack_paper is cleaned via ON DELETE CASCADE.
    DELETE FROM public.stack AS s
    WHERE s.deleted_at IS NOT NULL
      AND s.deleted_at < v_cutoff;

END;
$$;

COMMENT ON PROCEDURE
public.cleanup_soft_deleted_entries()
IS
'Permanently deletes papers and stacks whose deleted_at timestamp is older than the fixed 30 day retention period.';

-- Override the routine grants inherited from the default privileges in 001_init.sql.
REVOKE ALL
ON PROCEDURE public.cleanup_soft_deleted_entries()
FROM PUBLIC, app_rw, app_ro;

RESET ROLE;

COMMIT;
