#!/bin/sh
set -eu

PGPASSWORD="$(cat /run/secrets/db_backup_password)"
export PGPASSWORD
if ! mkdir /backups/.backup-lock; then
  echo "Another backup is running or a stale lock exists" >&2
  exit 1
fi
trap 'rmdir /backups/.backup-lock' EXIT

incremental_arg=""
case "${1:-full}" in
full)
  ;;
incremental)
  if [ ! -f /backups/latest/backup_manifest ]; then
    echo "Create a full backup first" >&2
    exit 1
  fi
  incremental_arg="--incremental=/backups/latest/backup_manifest"
  ;;
*)
  echo "Usage: backup.sh [full|incremental]" >&2
  exit 1
  ;;
esac

backup_name="$(date -u +%Y%m%dT%H%M%SZ)"
backup_dir="/backups/$backup_name"
test ! -e "$backup_dir"
mkdir "$backup_dir.partial"

pg_basebackup --host=db --username=paperstacks_backup --no-password \
  --pgdata="$backup_dir.partial" --format=tar --compress=gzip \
  ${incremental_arg:+"$incremental_arg"} \
  --wal-method=stream --checkpoint=fast --progress
pg_verifybackup --no-parse-wal "$backup_dir.partial"

mv "$backup_dir.partial" "$backup_dir"
ln -sfn "$backup_name" /backups/latest
echo "Verified backup: $backup_dir"
