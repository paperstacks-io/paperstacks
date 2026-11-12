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

# Barman cron maintains the WAL receiver and processes received segments.
# Base backups are triggered separately.
while true; do
  if ! barman -c /etc/paperstacks/barman.conf cron; then
    echo "Barman maintenance failed; retrying in 30 seconds" >&2
  fi
  sleep 30
done
