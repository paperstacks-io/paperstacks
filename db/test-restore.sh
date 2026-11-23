#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_dir="$(dirname "$script_dir")"
project="paperstacks-restore-test-$(date +%s)-$$"
backup_volume="$project-backups"
full_volume="$project-full"
chain_volume="$project-chain"
invalid_volume="$project-invalid"
work_dir="$(mktemp -d)"
export COMPOSE_PROJECT_NAME="$project"
export COMPOSE_FILE="$project_dir/compose.yaml:$work_dir/compose.yaml"
database_service=db

cleanup() {
    local status=$?
    if [ "$status" -ne 0 ]; then
        docker compose --profile restore logs db db-restored >&2 || true
    fi
    docker compose --profile restore down --volumes >/dev/null 2>&1 || true
    docker volume rm "$backup_volume" "$full_volume" "$chain_volume" "$invalid_volume" >/dev/null 2>&1 || true
    rm -rf "$work_dir"
}
trap cleanup EXIT

command -v make >/dev/null
docker info >/dev/null
for secret in db_super_password db_app_rw_password db_app_ro_password db_backup_password; do
    printf '%s' 'restore-test-password' > "$work_dir/$secret"
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
  db_backup_password:
    file: $work_dir/db_backup_password
EOF

sql() {
    docker compose exec -T --user postgres "$database_service" psql -XAt -v ON_ERROR_STOP=1 -U postgres -d paperstacks "$@"
}
wait_ready() {
    for attempt in {1..120}; do
        if docker compose exec -T "$database_service" pg_isready -h 127.0.0.1 -U postgres >/dev/null 2>&1 &&
            sql -c 'SELECT 1' >/dev/null 2>&1; then
            return
        fi
        sleep 1
    done
    echo "PostgreSQL did not become ready" >&2
    return 1
}
latest() {
    docker compose run --rm --no-deps -T --entrypoint sh \
        -v "$backup_volume:/backups:ro" db -c 'readlink /backups/latest'
}
restore_backup() {
    make -C "$project_dir" db-restore BACKUP="$1" \
        BACKUP_VOLUME="$backup_volume" RESTORE_VOLUME="$2"
}
start_restored() {
    export RESTORE_VOLUME="$1"
    database_service=db-restored
    docker compose --profile restore up -d --no-deps db-restored
    wait_ready
    [[ $(sql -c 'SHOW cron.launch_active_jobs') == off ]]
}

docker compose up -d --no-deps db
wait_ready
sql <<'SQL'
CREATE TABLE public.restore_test (id integer PRIMARY KEY, value text NOT NULL);
INSERT INTO public.restore_test VALUES (1, 'full');
SQL
make -C "$project_dir" db-backup-full BACKUP_VOLUME="$backup_volume"
full=$(latest)
sql -c "UPDATE public.restore_test SET value = 'incremental-1' WHERE id = 1"
sleep 1
make -C "$project_dir" db-backup-incremental BACKUP_VOLUME="$backup_volume"
incremental_one=$(latest)
sql -c "UPDATE public.restore_test SET value = 'incremental-2' WHERE id = 1"
sleep 1
make -C "$project_dir" db-backup-incremental BACKUP_VOLUME="$backup_volume"
incremental_two=$(latest)
sql -c "UPDATE public.restore_test SET value = 'after-backups' WHERE id = 1"

echo "Checking rejected missing input and missing backup"
if restore_backup "" "$invalid_volume" > "$work_dir/invalid.log" 2>&1; then exit 1; fi
grep -q 'Set BACKUP' "$work_dir/invalid.log"
if restore_backup "does-not-exist" "$invalid_volume" > "$work_dir/invalid.log" 2>&1; then exit 1; fi
grep -q 'Missing backup manifest' "$work_dir/invalid.log"

echo "Restoring the full backup while the source is still running"
restore_backup "$full" "$full_volume"
[[ $(sql -c 'SELECT value FROM public.restore_test WHERE id = 1') == after-backups ]]
if restore_backup "$full" "$full_volume" > "$work_dir/invalid.log" 2>&1; then exit 1; fi
grep -q 'Restore volume must be empty' "$work_dir/invalid.log"

echo "Combining a full backup and two incremental backups"
restore_backup "$full $incremental_one $incremental_two" "$chain_volume"
if restore_backup "$full $incremental_two" "$invalid_volume" > "$work_dir/invalid.log" 2>&1; then
    echo "An incomplete chain unexpectedly passed" >&2
    exit 1
fi

docker compose stop db
start_restored "$full_volume"
[[ $(sql -c 'SELECT value FROM public.restore_test WHERE id = 1') == full ]]
docker compose stop db-restored
start_restored "$chain_volume"
[[ $(sql -c 'SELECT value FROM public.restore_test WHERE id = 1') == incremental-2 ]]
sleep 1
make -C "$project_dir" db-backup-full BACKUP_VOLUME="$backup_volume" BACKUP_HOST=db-restored

echo "PASS: full and chained restores, rejected invalid inputs, preserved source, disabled scheduled jobs, backup of restored database"
