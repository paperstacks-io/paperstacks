\getenv backup_password PAPERSTACKS_BACKUP_PASSWORD
BEGIN;
SELECT 'CREATE ROLE paperstacks_backup'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'paperstacks_backup')
\gexec
ALTER ROLE paperstacks_backup LOGIN REPLICATION NOSUPERUSER NOCREATEDB NOCREATEROLE
  PASSWORD :'backup_password';
GRANT CONNECT ON DATABASE postgres TO paperstacks_backup;
GRANT pg_monitor TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_backup_start(text, boolean) TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_backup_stop(boolean) TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_switch_wal() TO paperstacks_backup;
GRANT EXECUTE ON FUNCTION pg_catalog.pg_create_restore_point(text) TO paperstacks_backup;
COMMIT;
