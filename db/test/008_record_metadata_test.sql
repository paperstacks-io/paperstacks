\set ON_ERROR_STOP on
\connect paperstacks

BEGIN;
SET LOCAL ROLE app_rw;

INSERT INTO public.app_user (external_id, email) VALUES
    ('metadata-owner', 'metadata-owner@example.invalid'),
    ('metadata-author', 'metadata-author@example.invalid'),
    ('metadata-editor', 'metadata-editor@example.invalid');

INSERT INTO public.paper (uuid, title, created_by, updated_by)
VALUES ('62000000-0000-0000-0000-000000000001', 'Metadata test', 'metadata-author', 'metadata-author');
INSERT INTO public.stack (uuid, name, owner_external_id, created_by, updated_by)
VALUES ('62000000-0000-0000-0000-000000000002', 'Metadata test', 'metadata-owner', 'metadata-author', 'metadata-author');

-- Save initial timestamps to verify creation metadata survives updates.
CREATE TEMP TABLE metadata_initial ON COMMIT DROP AS
SELECT 'paper' AS entity, created_at, updated_at FROM public.paper
WHERE uuid = '62000000-0000-0000-0000-000000000001'
UNION ALL
SELECT 'stack', created_at, updated_at FROM public.stack
WHERE uuid = '62000000-0000-0000-0000-000000000002';

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM metadata_initial WHERE created_at IS NULL OR updated_at IS DISTINCT FROM created_at) THEN
        RAISE EXCEPTION 'Creation and initial update timestamps must be populated equally';
    END IF;
END
$$;

-- Distinct statements in the same transaction must have distinct update times.
SELECT pg_sleep(0.01);
UPDATE public.paper SET title = 'Edited', updated_by = 'metadata-editor', updated_at = '2000-01-01Z'
WHERE uuid = '62000000-0000-0000-0000-000000000001';
UPDATE public.stack SET name = 'Edited', updated_by = 'metadata-editor', updated_at = '2000-01-01Z'
WHERE uuid = '62000000-0000-0000-0000-000000000002';

DO $$
DECLARE
    v_entity text;
    row_data record;
BEGIN
    FOREACH v_entity IN ARRAY ARRAY['paper', 'stack'] LOOP
        EXECUTE format('SELECT * FROM public.%I WHERE created_by = $1', v_entity)
            INTO STRICT row_data USING 'metadata-author';
        IF row_data.created_at IS DISTINCT FROM (SELECT i.created_at FROM metadata_initial i WHERE i.entity = v_entity)
           OR row_data.updated_at <= (SELECT i.updated_at FROM metadata_initial i WHERE i.entity = v_entity)
           OR row_data.updated_by IS DISTINCT FROM 'metadata-editor' THEN
            RAISE EXCEPTION 'Incorrect update metadata for %', v_entity;
        END IF;

        BEGIN
            EXECUTE format('UPDATE public.%I SET updated_by = $1 WHERE created_by = $2', v_entity)
                USING 'metadata-missing-user', 'metadata-author';
            RAISE EXCEPTION 'Unknown editor accepted for %', v_entity;
        EXCEPTION WHEN foreign_key_violation THEN
            NULL;
        END;
        BEGIN
            EXECUTE format('UPDATE public.%I SET created_by = $1 WHERE created_by = $2', v_entity)
                USING 'metadata-missing-user', 'metadata-author';
            RAISE EXCEPTION 'Unknown creator accepted for %', v_entity;
        EXCEPTION WHEN foreign_key_violation THEN
            NULL;
        END;
    END LOOP;
END
$$;

-- External identity changes cascade; deleting an actor preserves the records.
UPDATE public.app_user SET external_id = 'metadata-editor-renamed' WHERE external_id = 'metadata-editor';
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM public.paper WHERE updated_by = 'metadata-editor-renamed')
       OR NOT EXISTS (SELECT 1 FROM public.stack WHERE updated_by = 'metadata-editor-renamed') THEN
        RAISE EXCEPTION 'Editor identity update did not cascade';
    END IF;
END
$$;
DELETE FROM public.app_user WHERE external_id IN ('metadata-author', 'metadata-editor-renamed');
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM public.paper WHERE uuid = '62000000-0000-0000-0000-000000000001' AND created_by IS NULL AND updated_by IS NULL)
       OR NOT EXISTS (SELECT 1 FROM public.stack WHERE uuid = '62000000-0000-0000-0000-000000000002' AND created_by IS NULL AND updated_by IS NULL) THEN
        RAISE EXCEPTION 'Deleting actors must clear references without deleting records';
    END IF;
END
$$;

-- Existing callers and system imports may omit actor IDs.
INSERT INTO public.paper (uuid, title) VALUES ('62000000-0000-0000-0000-000000000003', 'System import');
INSERT INTO public.stack (uuid, name, owner_external_id)
VALUES ('62000000-0000-0000-0000-000000000004', 'System stack', 'metadata-owner');

-- Session timezone changes affect rendering, not the stored instant.
SET LOCAL TIME ZONE 'UTC';
CREATE TEMP TABLE metadata_epoch ON COMMIT DROP AS
SELECT extract(epoch FROM created_at) AS instant FROM public.paper WHERE uuid = '62000000-0000-0000-0000-000000000001';
SET LOCAL TIME ZONE 'Europe/Berlin';
DO $$
BEGIN
    IF (SELECT extract(epoch FROM created_at) FROM public.paper WHERE uuid = '62000000-0000-0000-0000-000000000001')
       IS DISTINCT FROM (SELECT instant FROM metadata_epoch) THEN
        RAISE EXCEPTION 'Timezone changed the stored instant';
    END IF;
END
$$;

ROLLBACK;
\echo 'Record metadata tests passed.'
