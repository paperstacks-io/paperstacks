#!/bin/sh
set -eu

mode="${1:-full}"
case "$mode" in
    full|incremental) ;;
    *)
        echo "Usage: backup.sh [full|incremental]" >&2
        exit 1
        ;;
esac

export PGPASSWORD="$(cat /run/secrets/db_backup_password)"
if ! mkdir /backups/.backup-lock; then
    echo "Another backup is running or a stale lock exists" >&2
    exit 1
fi
trap 'rmdir /backups/.backup-lock' EXIT

set --
if [ "$mode" = incremental ]; then
    if [ ! -f /backups/latest/backup_manifest ]; then
        echo "Create a full backup first" >&2
        exit 1
    fi
    set -- --incremental=/backups/latest/backup_manifest
fi

backup_dir="/backups/$(date -u +%Y%m%dT%H%M%SZ)"
test ! -e "$backup_dir"
mkdir "$backup_dir.partial"

pg_basebackup --host="${PGHOST:-db}" --username=paperstacks_backup --no-password \
    --pgdata="$backup_dir.partial" --format=tar --compress=gzip \
    "$@" --wal-method=stream --checkpoint=fast --progress
pg_verifybackup --no-parse-wal "$backup_dir.partial"

mv "$backup_dir.partial" "$backup_dir"
ln -sfn "$(basename "$backup_dir")" /backups/latest
echo "Verified backup: $backup_dir"
