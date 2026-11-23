#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_dir="$(dirname "$script_dir")"
project="paperstacks-backup-test-$(date +%s)-$$"
backup_volume="${project}-backups"
restore_volume="${project}-restore"
restored="${project}-restored"
image="${project}-postgres"
work_dir="$(mktemp -d)"
export COMPOSE_PROJECT_NAME="$project"
export COMPOSE_FILE="$project_dir/compose.yaml:$work_dir/compose.yaml"

cleanup() {
  local status=$?
  if [ "$status" -ne 0 ]; then
    docker compose logs db >&2 || true
    docker logs "$restored" >&2 || true
  fi
  docker rm -f -v "$restored" >/dev/null 2>&1 || true
  docker compose down --volumes >/dev/null 2>&1 || true
  docker volume rm "$backup_volume" "$restore_volume" >/dev/null 2>&1 || true
  docker image rm "$image" >/dev/null 2>&1 || true
  rm -rf "$work_dir"
}
trap cleanup EXIT

command -v make >/dev/null
docker info >/dev/null
for secret in db_super_password db_app_rw_password db_app_ro_password; do
  printf '%s' "backup-test-password" > "$work_dir/$secret"
done
cat > "$work_dir/compose.yaml" <<EOF
services:
  db:
    ports: !reset []
secrets:
  db_super_password:
    file: $work_dir/db_super_password
  db_app_rw_password:
    file: $work_dir/db_app_rw_password
  db_app_ro_password:
    file: $work_dir/db_app_ro_password
EOF

sql() {
  docker compose exec -T --user postgres db psql -XAt -v ON_ERROR_STOP=1 -U postgres -d paperstacks "$@"
}
wait_ready() {
  for attempt in {1..120}; do
    if "$@" >/dev/null 2>&1; then return; fi
    sleep 1
  done
  echo "Timed out waiting for PostgreSQL" >&2
  return 1
}
backup_files() {
  docker compose run --rm --no-deps -T --entrypoint sh \
    -v "$backup_volume:/backups:ro" db -c 'ls -1 /backups'
}

echo "Starting isolated source database"
docker compose up -d --no-deps db
wait_ready docker compose exec -T db pg_isready -h 127.0.0.1 -U postgres -d paperstacks
wait_ready sql -c 'SELECT 1'
sql <<'SQL'
CREATE TABLE public.backup_test (id integer PRIMARY KEY, value text NOT NULL);
INSERT INTO public.backup_test VALUES (1, 'before-backup');
SQL

echo "Creating the first backup through Make"
make -C "$project_dir" db-backup-full BACKUP_VOLUME="$backup_volume" > "$work_dir/backup.log" 2>&1 &
backup_pid=$!
activity=0
while kill -0 "$backup_pid" 2>/dev/null; do
  sql -c "INSERT INTO public.backup_test VALUES (2, 'online') ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value; SELECT value FROM public.backup_test WHERE id = 1;" >/dev/null
  activity=$((activity + 1))
  sleep 0.2
done
if ! wait "$backup_pid"; then cat "$work_dir/backup.log" >&2; exit 1; fi
cat "$work_dir/backup.log"
[[ $activity -gt 0 ]]
first_backup=$(backup_files)
[[ $(printf '%s\n' "$first_backup" | wc -l) -eq 1 ]]
sql -c "UPDATE public.backup_test SET value = 'after-backup' WHERE id = 1"

echo "Creating a second versioned backup"
sleep 1
make -C "$project_dir" db-backup-full BACKUP_VOLUME="$backup_volume"
[[ $(backup_files | wc -l) -eq 2 ]]

# Reuse the source container's installed pg_cron library for the restore.
source_id=$(docker compose ps -q db)
docker commit "$source_id" "$image" >/dev/null
docker compose stop db
docker run --rm --entrypoint sh -v "$backup_volume:/backups:ro" \
  -v "$restore_volume:/restore" "$image" -eu -c '
    cp -a "/backups/$1/." /restore/
    chown -R postgres:postgres /restore
    chmod 700 /restore
  ' sh "$first_backup"
docker run -d --name "$restored" --network none --entrypoint docker-entrypoint.sh \
  -e PGDATA=/restore -v "$restore_volume:/restore" "$image" \
  postgres -c shared_preload_libraries=pg_cron -c cron.database_name=paperstacks \
  -c cron.launch_active_jobs=off >/dev/null
wait_ready docker exec --user postgres "$restored" psql -XAt -U postgres -d paperstacks -c 'SELECT 1'
actual=$(docker exec --user postgres "$restored" psql -XAt -U postgres -d paperstacks \
  -c 'SELECT value FROM public.backup_test WHERE id = 1')
[[ $actual == before-backup ]]
echo "PASS: two verified backups, online reads/writes, restore with source stopped, original data recovered"
