-- Examples for the soft deletion mechanism introduced in 003_soft_delete.sql.
-- Replace the UUID placeholders with actual IDs.

-- ---------------------------------------------------------------------------
-- SOFT DELETE
-- ---------------------------------------------------------------------------

-- Soft-delete a paper.
-- The extra predicate makes the operation idempotent and preserves the
-- timestamp of the first deletion.
UPDATE public.paper
SET deleted_at = now()
WHERE uuid = '00000000-0000-0000-0000-000000000001'
  AND deleted_at IS NULL;

-- Soft-delete a stack.
UPDATE public.stack
SET deleted_at = now(),
    updated_at = now()
WHERE uuid = '00000000-0000-0000-0000-000000000002'
  AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- RESTORE
-- ---------------------------------------------------------------------------

-- Restore a paper.
UPDATE public.paper
SET deleted_at = NULL
WHERE uuid = '00000000-0000-0000-0000-000000000001'
  AND deleted_at IS NOT NULL;

-- Restore a stack.
UPDATE public.stack
SET deleted_at = NULL,
    updated_at = now()
WHERE uuid = '00000000-0000-0000-0000-000000000002'
  AND deleted_at IS NOT NULL;

-- ---------------------------------------------------------------------------
-- EXCLUDE SOFT-DELETED ENTRIES
-- ---------------------------------------------------------------------------

-- All active papers.
SELECT *
FROM public.paper
WHERE deleted_at IS NULL;

-- One active paper by UUID.
SELECT *
FROM public.paper
WHERE uuid = '00000000-0000-0000-0000-000000000001'
  AND deleted_at IS NULL;

-- All active stacks.
SELECT *
FROM public.stack
WHERE deleted_at IS NULL;

-- All active stacks owned by one user.
SELECT *
FROM public.stack
WHERE owner_external_id = 'example-user'
  AND deleted_at IS NULL;

-- Active papers contained in an active stack.
-- Both entities are checked so neither a deleted stack nor a deleted paper is
-- exposed by the relation table.
SELECT p.*
FROM public.stack_paper AS sp
JOIN public.stack AS s
  ON s.uuid = sp.uuid_stack
JOIN public.paper AS p
  ON p.uuid = sp.uuid_paper
WHERE s.uuid = '00000000-0000-0000-0000-000000000002'
  AND s.deleted_at IS NULL
  AND p.deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- OPTIONAL PERMANENT CLEANUP / RETENTION
-- ---------------------------------------------------------------------------
-- PostgreSQL VACUUM/autovacuum does NOT remove rows merely because deleted_at
-- is set. The rows are still live tuples. A scheduled application job,
-- system cron/Kubernetes CronJob, or an installed scheduling extension such as
-- pg_cron is needed to issue the final DELETE after the chosen retention time.
--
-- Example: permanently delete stacks that have been soft-deleted for 30 days.
-- stack_paper rows are removed through the existing ON DELETE CASCADE FK.
--
-- DELETE FROM public.stack
-- WHERE deleted_at < now() - interval '30 days';
--
-- For papers, paper_author currently uses ON DELETE RESTRICT. Therefore those
-- associations must be removed before the final paper DELETE. stack_paper,
-- pdf and metadata are already configured with ON DELETE CASCADE.
--
-- BEGIN;
-- DELETE FROM public.paper_author AS pa
-- USING public.paper AS p
-- WHERE pa.uuid_paper = p.uuid
--   AND p.deleted_at < now() - interval '30 days';
--
-- DELETE FROM public.paper
-- WHERE deleted_at < now() - interval '30 days';
-- COMMIT;
