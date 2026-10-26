\set ON_ERROR_STOP on

\connect paperstacks

-- Integration test for 005_create_soft_delete_cleanup_procedure.sql.
-- It verifies both the pg_cron registration and the hard-cleanup behavior.
-- All test data is rolled back at the end.

-- Run as postgres, the login that registers and executes the cron job.
-- Verify the schedule, database, command and execution role.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM cron.job
        WHERE jobname = 'soft-delete-cleanup'
          AND schedule = '0 3 * * *'
          AND database = 'paperstacks'
          AND command = 'CALL public.cleanup_soft_deleted_entries();'
          AND username = current_user
          AND active = true
    ) THEN
        RAISE EXCEPTION 'Test failed: soft-delete-cleanup cron job is missing or misconfigured';
    END IF;
END
$$;

BEGIN;

DO $$
BEGIN
    IF to_regprocedure('public.cleanup_soft_deleted_entries(interval)') IS NOT NULL THEN
        RAISE EXCEPTION 'Test failed: configurable retention overload still exists';
    END IF;
    IF has_function_privilege('app_rw', 'public.cleanup_soft_deleted_entries()', 'EXECUTE')
       OR has_function_privilege('app_ro', 'public.cleanup_soft_deleted_entries()', 'EXECUTE') THEN
        RAISE EXCEPTION 'Test failed: application roles can execute administrative cleanup';
    END IF;
END
$$;

-- One expired pair (>30 days) and one recent pair (<30 days).
INSERT INTO public.app_user (external_id, email)
VALUES ('cleanup-test-user', 'cleanup-test@example.invalid');

INSERT INTO public.paper (uuid, doi, title, pdf_storage_key, deleted_at)
VALUES
(
    '33333333-3333-3333-3333-333333333333',
    '10.0000/cleanup-expired',
    'Expired cleanup test paper',
    'pdf/33333333-3333-3333-3333-333333333333.pdf',
    now() - interval '31 days'
),
(
    '44444444-4444-4444-4444-444444444444',
    '10.0000/cleanup-recent',
    'Recent cleanup test paper',
    'pdf/44444444-4444-4444-4444-444444444444.pdf',
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
-- existing ON DELETE CASCADE behavior (metadata).
DO $$
DECLARE
    v_author_key bigint;
BEGIN
    INSERT INTO public.author (name_first)
    VALUES ('Cleanup Test Author')
    RETURNING key INTO v_author_key;

    INSERT INTO public.paper_author (uuid_paper, key_author, position)
    VALUES
        ('33333333-3333-3333-3333-333333333333', v_author_key, 1),
        ('44444444-4444-4444-4444-444444444444', v_author_key, 1);
END
$$;

INSERT INTO public.metadata (publisher, uuid_paper)
VALUES
    ('Expired Publisher', '33333333-3333-3333-3333-333333333333'),
    ('Recent Publisher', '44444444-4444-4444-4444-444444444444');

-- Call through the same postgres login used by pg_cron. SECURITY DEFINER causes the
-- cleanup body to execute with app_owner privileges.
CALL public.cleanup_soft_deleted_entries();

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
          AND pdf_storage_key = 'pdf/44444444-4444-4444-4444-444444444444.pdf'
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

-- Keep setup and CALL in one statement so the exact cutoff cannot drift.
DO $$
DECLARE
    boundary_paper uuid := '77777777-7777-7777-7777-777777777777';
    boundary_stack uuid := '88888888-8888-8888-8888-888888888888';
    active_paper uuid := '99999999-9999-9999-9999-999999999999';
    active_stack uuid := 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
BEGIN
    INSERT INTO public.paper (uuid, doi, title, deleted_at)
    VALUES
        (boundary_paper, '10.0000/cleanup-boundary', 'Boundary paper', statement_timestamp() - interval '30 days'),
        (active_paper, '10.0000/cleanup-active', 'Active paper', NULL);
    INSERT INTO public.stack (uuid, name, owner_external_id, deleted_at)
    VALUES
        (boundary_stack, 'Boundary stack', 'cleanup-test-user', statement_timestamp() - interval '30 days'),
        (active_stack, 'Active stack', 'cleanup-test-user', NULL);

    CALL public.cleanup_soft_deleted_entries();

    IF (SELECT count(*) FROM public.paper WHERE uuid IN (boundary_paper, active_paper)) <> 2
       OR (SELECT count(*) FROM public.stack WHERE uuid IN (boundary_stack, active_stack)) <> 2 THEN
        RAISE EXCEPTION 'Test failed: active entities or entities at the exact 30 day cutoff were removed';
    END IF;
END
$$;

ROLLBACK;
RESET ROLE;

\echo 'Soft-delete cleanup tests passed.'
