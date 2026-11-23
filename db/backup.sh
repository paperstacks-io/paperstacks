#!/bin/sh
set -eu

mode="${1:-full}"
if [ "$mode" != full ] && [ "$mode" != incremental ]; then
  echo "Usage: backup.sh [full|incremental]" >&2
  exit 1
fi

PGPASSWORD="$(cat /run/secrets/db_backup_password)"
export PGPASSWORD
if ! mkdir /backups/.backup-lock; then
  echo "Another backup is running or a stale lock exists" >&2
  exit 1
fi
trap 'rmdir /backups/.backup-lock' EXIT

incremental_arg=""
if [ "$mode" = incremental ]; then
  if [ ! -f /backups/latest/backup_manifest ]; then
    echo "Create a full backup first" >&2
    exit 1
  fi
  incremental_arg="--incremental=/backups/latest/backup_manifest"
fi

# Perform the backup
mkdir /tmp/backup

pg_basebackup --host=db --username=paperstacks_backup --no-password \
  --pgdata=/tmp/backup --format=tar --compress=gzip \
  ${incremental_arg:+"$incremental_arg"} \
  --wal-method=stream --checkpoint=fast --progress

pg_verifybackup --no-parse-wal /tmp/backup

# Transfer the backup to the host only after verification succeeds.
backup_name="$(date -u +%Y-%m-%d_T%H%M%SZ)-$mode"
backup_dir="/backups/$backup_name"
if [ -e "$backup_dir" ]; then
  echo "Backup destination already exists: $backup_dir" >&2
  exit 1
fi
mv /tmp/backup "$backup_dir"
ln -sfn "$backup_name" /backups/latest
echo "Verified backup: $backup_dir"
