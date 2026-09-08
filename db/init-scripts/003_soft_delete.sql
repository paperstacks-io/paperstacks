\set ON_ERROR_STOP on

\connect paperstacks

BEGIN;

SET ROLE app_owner;
SET search_path TO public;

-- Soft deletion for papers and stacks.
-- NULL      = active
-- NOT NULL  = soft deleted at the stored timestamp
ALTER TABLE public.paper
    ADD COLUMN deleted_at timestamptz NULL;

ALTER TABLE public.stack
    ADD COLUMN deleted_at timestamptz NULL;

COMMENT ON COLUMN public.paper.deleted_at IS
'Soft deletion timestamp. NULL means active; non-NULL means deleted.';

COMMENT ON COLUMN public.stack.deleted_at IS
'Soft deletion timestamp. NULL means active; non-NULL means deleted.';

-- These partial indexes are intended for administrative queries and a future
-- retention-based hard cleanup. Active-row lookups still use the existing
-- primary/unique indexes and explicitly filter with deleted_at IS NULL.
CREATE INDEX paper_deleted_at_idx
    ON public.paper (deleted_at)
    WHERE deleted_at IS NOT NULL;

CREATE INDEX stack_deleted_at_idx
    ON public.stack (deleted_at)
    WHERE deleted_at IS NOT NULL;

COMMIT;
