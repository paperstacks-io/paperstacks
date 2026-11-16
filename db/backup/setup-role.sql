\getenv backup_password PAPERSTACKS_BACKUP_PASSWORD
BEGIN;
SELECT 'CREATE ROLE paperstacks_backup'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'paperstacks_backup')
\gexec
ALTER ROLE paperstacks_backup LOGIN REPLICATION NOSUPERUSER NOCREATEDB NOCREATEROLE
  PASSWORD :'backup_password';
GRANT CONNECT ON DATABASE postgres TO paperstacks_backup;
GRANT pg_monitor TO paperstacks_backup;
GRANT pg_checkpoint TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_backup_start(text, boolean) TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_backup_stop(boolean) TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_switch_wal() TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_create_restore_point(text) TO paperstacks_backup;

CREATE SCHEMA IF NOT EXISTS backup_jobs;
CREATE TABLE IF NOT EXISTS backup_jobs.weekly_full (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    requested bigint NOT NULL DEFAULT 1,
    completed bigint NOT NULL DEFAULT 0
);
INSERT INTO backup_jobs.weekly_full DEFAULT VALUES ON CONFLICT DO NOTHING;
GRANT USAGE ON SCHEMA backup_jobs TO paperstacks_backup;
GRANT SELECT, UPDATE (completed) ON backup_jobs.weekly_full TO paperstacks_backup;
COMMIT;

\connect paperstacks
SELECT cron.schedule_in_database(
    'barman-weekly-full',
    '0 2 * * 0',
    $$UPDATE backup_jobs.weekly_full SET requested = requested + 1;$$,
    'postgres'
);
