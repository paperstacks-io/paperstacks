#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

compose_file="$script_dir/compose.yaml"
compose_project="paperstacks-db-smoke"
secrets_dir="$(mktemp -d)"
compose_override="$secrets_dir/compose.smoke.yaml"
compose_cmd=()
compose_args=()
compose_started=false

super_password="paperstacks-super-smoke"
app_rw_password="paperstacks-rw-smoke"
app_ro_password="paperstacks-ro-smoke"

on_exit() {
  status=$?
  if [ "$status" -ne 0 ] && [ "$compose_started" = true ]; then
    echo "Postgres init smoke test failed; container logs follow:" >&2
    "${compose_cmd[@]}" "${compose_args[@]}" logs db >&2 || true
  fi
  if [ "$compose_started" = true ]; then
    "${compose_cmd[@]}" "${compose_args[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  fi
  rm -rf "$secrets_dir"
  exit "$status"
}
trap on_exit EXIT

select_compose_command() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    compose_cmd=(docker compose)
  elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
    compose_cmd=(podman compose)
  else
    echo "docker compose or podman compose is required for the DB init smoke test" >&2
    exit 1
  fi
}

wait_for_postgres() {
  for _ in {1..60}; do
    if "${compose_cmd[@]}" "${compose_args[@]}" exec -T db pg_isready -h 127.0.0.1 -U postgres -d postgres >/dev/null 2>&1; then
      return 0
    fi

    if [ -z "$("${compose_cmd[@]}" "${compose_args[@]}" ps --status running -q db)" ]; then
      echo "Postgres container exited before becoming ready" >&2
      return 1
    fi

    sleep 1
  done

  echo "Timed out waiting for Postgres to become ready" >&2
  return 1
}

psql_as() {
  local user="$1"
  local password="$2"
  local database="$3"
  shift 3

  "${compose_cmd[@]}" "${compose_args[@]}" exec -T \
    -e PGPASSWORD="$password" \
    db \
    psql \
    -h 127.0.0.1 \
    -U "$user" \
    -d "$database" \
    -v ON_ERROR_STOP=1 \
    -X \
    "$@"
}

select_compose_command
compose_args=(--project-name "$compose_project" --file "$compose_file" --file "$compose_override")

printf '%s\n' "$super_password" >"$secrets_dir/db_super_password"
printf '%s\n' "$app_rw_password" >"$secrets_dir/db_app_rw_password"
printf '%s\n' "$app_ro_password" >"$secrets_dir/db_app_ro_password"

cat >"$compose_override" <<EOF
services:
  db:
    ports: !reset []
secrets:
  db_super_password:
    file: $secrets_dir/db_super_password
  db_app_rw_password:
    file: $secrets_dir/db_app_rw_password
  db_app_ro_password:
    file: $secrets_dir/db_app_ro_password
EOF

compose_started=true
"${compose_cmd[@]}" "${compose_args[@]}" up -d --no-deps db >/dev/null

wait_for_postgres

psql_as postgres "$super_password" postgres -c \
  "DO \$\$
   BEGIN
     IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'paperstacks') THEN
       RAISE EXCEPTION 'database paperstacks was not created';
     END IF;

     IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_owner' AND NOT rolcanlogin) THEN
       RAISE EXCEPTION 'role app_owner was not created as NOLOGIN';
     END IF;

     IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_rw' AND rolcanlogin) THEN
       RAISE EXCEPTION 'login role app_rw was not created';
     END IF;

     IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_ro' AND rolcanlogin) THEN
       RAISE EXCEPTION 'login role app_ro was not created';
     END IF;
   END
   \$\$;"

psql_as postgres "$super_password" paperstacks -c \
  "DO \$\$
   DECLARE
     expected_table text;
   BEGIN
     FOREACH expected_table IN ARRAY ARRAY[
       'public.affiliation',
       'public.author',
       'public.paper',
       'public.paper_author',
       'public.pdf',
       'public.metadata'
     ] LOOP
       IF to_regclass(expected_table) IS NULL THEN
         RAISE EXCEPTION 'missing table %', expected_table;
       END IF;
     END LOOP;

     IF to_regtype('public.paper_type') IS NULL THEN
       RAISE EXCEPTION 'missing enum public.paper_type';
     END IF;

     IF to_regtype('public.publication_status') IS NULL THEN
       RAISE EXCEPTION 'missing enum public.publication_status';
     END IF;

     IF EXISTS (
       SELECT 1
       FROM pg_tables
       WHERE schemaname = 'public'
         AND tablename IN ('affiliation', 'author', 'paper', 'paper_author', 'pdf', 'metadata')
         AND tableowner <> 'app_owner'
     ) THEN
       RAISE EXCEPTION 'one or more application tables are not owned by app_owner';
     END IF;
   END
   \$\$;"

psql_as app_rw "$app_rw_password" paperstacks -c \
  "INSERT INTO public.affiliation (name) VALUES ('Smoke Test Lab');
   UPDATE public.affiliation SET name = 'Smoke Test Lab Updated' WHERE name = 'Smoke Test Lab';
   SELECT key FROM public.affiliation WHERE name = 'Smoke Test Lab Updated';"

psql_as app_ro "$app_ro_password" paperstacks -c \
  "SELECT key, name FROM public.affiliation WHERE name = 'Smoke Test Lab Updated';"

if psql_as app_ro "$app_ro_password" paperstacks -c \
  "INSERT INTO public.affiliation (name) VALUES ('Read Only Should Fail');" >/dev/null 2>&1; then
  echo "app_ro unexpectedly has INSERT privileges" >&2
  exit 1
fi

psql_as app_rw "$app_rw_password" paperstacks -c \
  "DELETE FROM public.affiliation WHERE name = 'Smoke Test Lab Updated';"

echo "DB init smoke test passed"
