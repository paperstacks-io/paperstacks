\set ON_ERROR_STOP on

\connect paperstacks

-- Integration test for 006_soft_delete_cleanup.sql.
-- It verifies both the pg_cron registration and the hard-cleanup behavior.
-- All test data is rolled back at the end.

SET ROLE app_rw;

-- Verify that this user can see the cron job it owns and that it has the
-- expected schedule/database.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM cron.job
        WHERE jobname = 'soft-delete-cleanup'
          AND schedule = '0 3 * * *'
          AND database = 'paperstacks'
          AND active = true
    ) THEN
        RAISE EXCEPTION 'Test failed: soft-delete-cleanup cron job is missing or misconfigured';
    END IF;
END
$$;

BEGIN;

-- One expired pair (>30 days) and one recent pair (<30 days).
INSERT INTO public.app_user (external_id, email)
VALUES ('cleanup-test-user', 'cleanup-test@example.invalid');

INSERT INTO public.paper (uuid, doi, title, deleted_at)
VALUES
(
    '33333333-3333-3333-3333-333333333333',
    '10.0000/cleanup-expired',
    'Expired cleanup test paper',
    now() - interval '31 days'
),
(
    '44444444-4444-4444-4444-444444444444',
    '10.0000/cleanup-recent',
    'Recent cleanup test paper',
    now() - interval '29 days'
);

INSERT INTO public.stack (uuid, name, owner_external_id, deleted_at)
VALUES
(
    '55555555-5555-5555-5555-555555555555',
    'Expired cleanup test stack',
    'cleanup-test-user',
    now() - interval '31 days'
),
(
    '66666666-6666-6666-6666-666666666666',
    'Recent cleanup test stack',
    'cleanup-test-user',
    now() - interval '29 days'
);

INSERT INTO public.stack_paper (uuid_stack, uuid_paper)
VALUES
(
    '55555555-5555-5555-5555-555555555555',
    '33333333-3333-3333-3333-333333333333'
),
(
    '66666666-6666-6666-6666-666666666666',
    '44444444-4444-4444-4444-444444444444'
);

-- Add dependent data to verify both explicit cleanup (paper_author) and
-- existing ON DELETE CASCADE behavior (pdf and metadata).
DO $$
DECLARE
    v_author_key bigint;
BEGIN
    INSERT INTO public.author (name_first)
    VALUES ('Cleanup Test Author')
    RETURNING key INTO v_author_key;

    INSERT INTO public.paper_author (uuid_paper, key_author)
    VALUES
        ('33333333-3333-3333-3333-333333333333', v_author_key),
        ('44444444-4444-4444-4444-444444444444', v_author_key);
END
$$;

INSERT INTO public.pdf (pdf_url, uuid_paper)
VALUES
    ('https://example.invalid/expired.pdf', '33333333-3333-3333-3333-333333333333'),
    ('https://example.invalid/recent.pdf', '44444444-4444-4444-4444-444444444444');

INSERT INTO public.metadata (publisher, uuid_paper)
VALUES
    ('Expired Publisher', '33333333-3333-3333-3333-333333333333'),
    ('Recent Publisher', '44444444-4444-4444-4444-444444444444');

-- Call through the same app role used by pg_cron. SECURITY DEFINER causes the
-- cleanup body to execute with app_owner privileges.
CALL public.cleanup_soft_deleted_entries(interval '30 days');

DO $$
BEGIN
    -- Expired entities must be permanently removed.
    IF EXISTS (
        SELECT 1 FROM public.paper
        WHERE uuid = '33333333-3333-3333-3333-333333333333'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired paper was not permanently deleted';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.stack
        WHERE uuid = '55555555-5555-5555-5555-555555555555'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired stack was not permanently deleted';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.paper_author
        WHERE uuid_paper = '33333333-3333-3333-3333-333333333333'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired paper_author relation still exists';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.stack_paper
        WHERE uuid_paper = '33333333-3333-3333-3333-333333333333'
           OR uuid_stack = '55555555-5555-5555-5555-555555555555'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired stack_paper relation still exists';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.pdf
        WHERE uuid_paper = '33333333-3333-3333-3333-333333333333'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired PDF still exists';
    END IF;

    IF EXISTS (
        SELECT 1 FROM public.metadata
        WHERE uuid_paper = '33333333-3333-3333-3333-333333333333'
    ) THEN
        RAISE EXCEPTION 'Test failed: expired metadata still exists';
    END IF;

    -- Recent soft-deleted entities must remain until the retention expires.
    IF NOT EXISTS (
        SELECT 1 FROM public.paper
        WHERE uuid = '44444444-4444-4444-4444-444444444444'
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: recent soft-deleted paper was removed too early';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM public.stack
        WHERE uuid = '66666666-6666-6666-6666-666666666666'
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: recent soft-deleted stack was removed too early';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM public.paper_author
        WHERE uuid_paper = '44444444-4444-4444-4444-444444444444'
    ) THEN
        RAISE EXCEPTION 'Test failed: recent paper_author relation was removed too early';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM public.stack_paper
        WHERE uuid_stack = '66666666-6666-6666-6666-666666666666'
          AND uuid_paper = '44444444-4444-4444-4444-444444444444'
    ) THEN
        RAISE EXCEPTION 'Test failed: recent stack_paper relation was removed too early';
    END IF;
END
$$;

ROLLBACK;
RESET ROLE;

\echo 'Soft-delete cleanup tests passed.'
