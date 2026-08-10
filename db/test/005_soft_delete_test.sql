\set ON_ERROR_STOP on

\connect paperstacks

-- This test runs inside a transaction and rolls everything back at the end.
-- It can therefore be executed against a development database without leaving
-- the test data behind.
BEGIN;

SET ROLE app_owner;
SET search_path TO public;

-- Fixed IDs keep the assertions simple and reproducible.
INSERT INTO public.app_user (external_id, email)
VALUES ('soft-delete-test-user', 'soft-delete-test@example.invalid');

INSERT INTO public.paper (uuid, doi, title)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    '10.0000/soft-delete-test',
    'Soft delete test paper'
);

INSERT INTO public.stack (uuid, name, owner_external_id)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'Soft delete test stack',
    'soft-delete-test-user'
);

INSERT INTO public.stack_paper (uuid_stack, uuid_paper)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111'
);

-- ---------------------------------------------------------------------------
-- PAPER: DELETE
-- ---------------------------------------------------------------------------
UPDATE public.paper
SET deleted_at = now()
WHERE uuid = '11111111-1111-1111-1111-111111111111'
  AND deleted_at IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = '11111111-1111-1111-1111-111111111111'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: deleted paper is still visible as active';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = '11111111-1111-1111-1111-111111111111'
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: paper was not marked as deleted';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = '22222222-2222-2222-2222-222222222222'
          AND uuid_paper = '11111111-1111-1111-1111-111111111111'
    ) THEN
        RAISE EXCEPTION 'Test failed: paper-stack relation was removed by soft delete';
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- PAPER: RESTORE
-- ---------------------------------------------------------------------------
UPDATE public.paper
SET deleted_at = NULL
WHERE uuid = '11111111-1111-1111-1111-111111111111'
  AND deleted_at IS NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = '11111111-1111-1111-1111-111111111111'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: paper was not restored';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = '22222222-2222-2222-2222-222222222222'
          AND uuid_paper = '11111111-1111-1111-1111-111111111111'
    ) THEN
        RAISE EXCEPTION 'Test failed: paper-stack relation was not preserved';
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- STACK: DELETE
-- ---------------------------------------------------------------------------
UPDATE public.stack
SET deleted_at = now(),
    updated_at = now()
WHERE uuid = '22222222-2222-2222-2222-222222222222'
  AND deleted_at IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = '22222222-2222-2222-2222-222222222222'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: deleted stack is still visible as active';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = '22222222-2222-2222-2222-222222222222'
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: stack was not marked as deleted';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = '22222222-2222-2222-2222-222222222222'
          AND uuid_paper = '11111111-1111-1111-1111-111111111111'
    ) THEN
        RAISE EXCEPTION 'Test failed: stack-paper relation was removed by soft delete';
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- STACK: RESTORE
-- ---------------------------------------------------------------------------
UPDATE public.stack
SET deleted_at = NULL,
    updated_at = now()
WHERE uuid = '22222222-2222-2222-2222-222222222222'
  AND deleted_at IS NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = '22222222-2222-2222-2222-222222222222'
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Test failed: stack was not restored';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = '22222222-2222-2222-2222-222222222222'
          AND uuid_paper = '11111111-1111-1111-1111-111111111111'
    ) THEN
        RAISE EXCEPTION 'Test failed: stack-paper relation was not preserved after restore';
    END IF;
END
$$;

ROLLBACK;

\echo 'Soft-delete tests passed.'
