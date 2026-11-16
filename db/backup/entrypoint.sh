#!/usr/bin/env bash
set -euo pipefail

umask 077
password=$(cat /run/secrets/db_barman_password)
if [[ ! $password =~ ^[a-zA-Z0-9_-]+$ ]]; then
  echo "Expected a nonempty URL-safe db_barman_password secret" >&2
  exit 1
fi
printf 'db:5432:*:paperstacks_backup:%s\n' "$password" > /tmp/paperstacks-barman.pgpass
unset password

if [ "$#" -gt 0 ]; then
  exec barman -c /etc/paperstacks/barman.conf "$@"
fi

# pg_cron requests a backup in PostgreSQL. Barman runs it here while WAL
# maintenance continues independently.
run_weekly_backups() {
  local request
  local conn='host=db port=5432 dbname=postgres user=paperstacks_backup passfile=/tmp/paperstacks-barman.pgpass connect_timeout=5'
  while true; do
    if request=$(psql -XAt -v ON_ERROR_STOP=1 "$conn" <<'SQL'
SELECT requested FROM backup_jobs.weekly_full
WHERE requested > completed
  AND EXISTS (SELECT FROM pg_replication_slots WHERE slot_name = 'paperstacks_barman' AND active);
SQL
    ); then
      if [[ $request =~ ^[0-9]+$ ]] &&
        barman -c /etc/paperstacks/barman.conf switch-wal --force --archive --archive-timeout 120 paperstacks &&
        barman -c /etc/paperstacks/barman.conf backup --wait paperstacks; then
        psql -X -v ON_ERROR_STOP=1 -v request="$request" "$conn" <<'SQL' ||
UPDATE backup_jobs.weekly_full SET completed = GREATEST(completed, :'request'::bigint);
SQL
          echo "Could not acknowledge the backup; the request remains pending" >&2
      fi
    fi
    sleep 60
  done
}

run_weekly_backups &

# Barman cron maintains the WAL receiver and processes received segments.
while true; do
  if ! barman -c /etc/paperstacks/barman.conf cron; then
    echo "Barman maintenance failed; retrying in 30 seconds" >&2
  fi
  sleep 30
done
