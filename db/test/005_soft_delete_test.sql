BEGIN;

DO $$
DECLARE
    test_user_id text :=
        'soft-delete-test-' ||
        substr(md5(random()::text || clock_timestamp()::text), 1, 12);

    test_paper_uuid uuid :=
        md5(random()::text || clock_timestamp()::text)::uuid;

    test_stack_uuid uuid :=
        md5(random()::text || clock_timestamp()::text)::uuid;
BEGIN

    ------------------------------------------------------------------
    -- SETUP
    ------------------------------------------------------------------

    INSERT INTO public.app_user (
        external_id,
        email
    )
    VALUES (
        test_user_id,
        test_user_id || '@example.test'
    );

    INSERT INTO public.paper (
        uuid,
        doi,
        title
    )
    VALUES (
        test_paper_uuid,
        '10.9999/' || test_user_id,
        'Soft Delete Test Paper'
    );

    INSERT INTO public.stack (
        uuid,
        name,
        owner_external_id
    )
    VALUES (
        test_stack_uuid,
        'Soft Delete Test Stack',
        test_user_id
    );

    INSERT INTO public.stack_paper (
        uuid_stack,
        uuid_paper
    )
    VALUES (
        test_stack_uuid,
        test_paper_uuid
    );


    ------------------------------------------------------------------
    -- TEST 1: PAPER SOFT DELETE
    ------------------------------------------------------------------

    UPDATE public.paper
    SET deleted_at = now()
    WHERE uuid = test_paper_uuid
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'TEST FAILED: Paper could not be soft deleted.';
    END IF;

    -- Paper must still physically exist.
    IF NOT EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = test_paper_uuid
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Deleted paper does not exist anymore.';
    END IF;

    -- Normal queries must no longer return the paper.
    IF EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = test_paper_uuid
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Soft-deleted paper is still visible.';
    END IF;

    -- stack_paper relation must still exist.
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = test_stack_uuid
          AND uuid_paper = test_paper_uuid
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: stack_paper relation was removed.';
    END IF;


    ------------------------------------------------------------------
    -- TEST 2: PAPER RESTORE
    ------------------------------------------------------------------

    UPDATE public.paper
    SET deleted_at = NULL
    WHERE uuid = test_paper_uuid
      AND deleted_at IS NOT NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'TEST FAILED: Paper could not be restored.';
    END IF;

    -- Paper must be visible again.
    IF NOT EXISTS (
        SELECT 1
        FROM public.paper
        WHERE uuid = test_paper_uuid
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Restored paper is not visible.';
    END IF;

    -- Relation must still exist after restore.
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = test_stack_uuid
          AND uuid_paper = test_paper_uuid
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: stack_paper relation is missing after restore.';
    END IF;


    ------------------------------------------------------------------
    -- TEST 3: STACK SOFT DELETE
    ------------------------------------------------------------------

    UPDATE public.stack
    SET deleted_at = now(),
        updated_at = now()
    WHERE uuid = test_stack_uuid
      AND deleted_at IS NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'TEST FAILED: Stack could not be soft deleted.';
    END IF;

    -- Stack must still physically exist.
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = test_stack_uuid
          AND deleted_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Deleted stack does not exist anymore.';
    END IF;

    -- Normal queries must not return the stack.
    IF EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = test_stack_uuid
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Soft-deleted stack is still visible.';
    END IF;

    -- Relation to the paper must remain.
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper
        WHERE uuid_stack = test_stack_uuid
          AND uuid_paper = test_paper_uuid
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: stack_paper relation was removed after stack deletion.';
    END IF;


    ------------------------------------------------------------------
    -- TEST 4: STACK RESTORE
    ------------------------------------------------------------------

    UPDATE public.stack
    SET deleted_at = NULL,
        updated_at = now()
    WHERE uuid = test_stack_uuid
      AND deleted_at IS NOT NULL;

    IF NOT FOUND THEN
        RAISE EXCEPTION
            'TEST FAILED: Stack could not be restored.';
    END IF;

    -- Stack must be visible again.
    IF NOT EXISTS (
        SELECT 1
        FROM public.stack
        WHERE uuid = test_stack_uuid
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Restored stack is not visible.';
    END IF;


    ------------------------------------------------------------------
    -- TEST 5: PAPER MUST BE VISIBLE IN RESTORED STACK
    ------------------------------------------------------------------

    IF NOT EXISTS (
        SELECT 1
        FROM public.stack_paper sp
        JOIN public.paper p
            ON p.uuid = sp.uuid_paper
        JOIN public.stack s
            ON s.uuid = sp.uuid_stack
        WHERE sp.uuid_stack = test_stack_uuid
          AND sp.uuid_paper = test_paper_uuid
          AND p.deleted_at IS NULL
          AND s.deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION
            'TEST FAILED: Paper is not visible in the restored stack.';
    END IF;


    ------------------------------------------------------------------
    -- SUCCESS
    ------------------------------------------------------------------

    RAISE NOTICE 'SUCCESS: All soft delete and restore tests passed.';

END
$$;

-- Remove all test data again.
ROLLBACK;