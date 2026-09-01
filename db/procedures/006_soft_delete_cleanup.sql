\set ON_ERROR_STOP on

\connect paperstacks

BEGIN;

SET ROLE app_owner;
SET search_path TO public;

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

    IF p_retention IS NULL
       OR p_retention < interval '0 seconds'
    THEN
        RAISE EXCEPTION
            'Retention period must be a non-negative interval';
    END IF;

    v_cutoff := statement_timestamp() - p_retention;

    -- paper_author uses ON DELETE RESTRICT.
    DELETE FROM public.paper_author AS pa
    USING public.paper AS p
    WHERE pa.uuid_paper = p.uuid
      AND p.deleted_at IS NOT NULL
      AND p.deleted_at < v_cutoff;

    -- stack_paper, pdf and metadata are cleaned via ON DELETE CASCADE.
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
public.cleanup_soft_deleted_entries(interval)
IS
'Permanently deletes papers and stacks whose deleted_at timestamp is older than the retention period.';

REVOKE ALL
ON PROCEDURE public.cleanup_soft_deleted_entries(interval)
FROM PUBLIC;

GRANT EXECUTE
ON PROCEDURE public.cleanup_soft_deleted_entries(interval)
TO app_rw;

RESET ROLE;

COMMIT;