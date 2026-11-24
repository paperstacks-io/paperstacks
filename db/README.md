# DB

## Run

Run `docker compose -f deploy/compose.yaml up -d db` from the repository root to
start the database. For a backend running on the host, use `compose.local.yaml`
to publish PostgreSQL on localhost port 5432:

```sh
docker compose -f deploy/compose.yaml -f compose.local.yaml up -d db
```

To reinitialize a disposable local database after changing the init scripts,
remove its volumes with `docker compose -f deploy/compose.yaml down -v` and
start the database again.


## Backend connection

Set `DATABASE_URL` with the `app_rw` user and omit the password. The server,
crawler, and Crossref sync read the password from `/secrets/db_app_rw_password.txt`
by default. When running directly from `backend/`, point
`DATABASE_PASSWORD_FILE` to the file created by `make secrets` at the repository
root:

```sh
export DATABASE_URL='postgres://app_rw@localhost:5432/paperstacks?sslmode=disable'
export DATABASE_PASSWORD_FILE='../secrets/db_app_rw_password.txt'
```

If `DATABASE_URL` is unset, the server uses in-memory repositories.

## Paper and stack record metadata

`002_init_database.sql` adds operational metadata directly to `paper` and
`stack`. It is separate from bibliographic `public.metadata` / `domain.Metadata`.

| Field | Type | Meaning |
| --- | --- | --- |
| `created_at` | `timestamptz NOT NULL` | Creation of the database record |
| `updated_at` | `timestamptz NOT NULL` | Latest SQL statement updating the row (LastUpdate) |
| `created_by` | nullable `text` | Creating user's `app_user.external_id` |
| `updated_by` | nullable `text` | Last editing user's `app_user.external_id` |

PostgreSQL stores `timestamptz` as an absolute instant with microsecond precision;
the session timezone controls its display. It does not retain the original
timezone name or offset. Use ISO 8601 with an explicit offset when supplying
timestamps and UTC (`Z`) when serializing them to clients. Defaults use
`statement_timestamp()`, so all rows written by a statement share one instant.
Unlike `now()`, this advances between statements within the same transaction.
These are statement times, not transaction commit times or publication dates.

On insertion both timestamps default to the statement start. On every row UPDATE,
a trigger overwrites `updated_at`, even for no-op updates and foreign-key actions.
Ordinary updates leave `created_at` intact; applications must not rewrite creation
metadata. Related-row changes (such as adding a `stack_paper` link) do not touch
the parent automatically: update the parent explicitly if that operation should
count as an edit. The init script creates this schema for a fresh database; it
does not migrate existing databases.

The application supplies both actor IDs on INSERT and `updated_by` on each UPDATE
(explicitly NULL for system/unknown actors). A shared SQL login such as `app_rw`
cannot identify the authenticated application user. Existing callers that omit
these fields remain compatible but do not provide actor attribution; the database
does not infer it. This schema change does not wire authentication into backend
writes. A stack's owner is independent of its creator/editor.

Actor IDs are foreign keys. Renaming an external ID cascades; deleting an actor
sets these references to NULL without deleting papers or stacks. Stack ownership
still has its existing RESTRICT constraint. NULL therefore means unknown, system,
or a deleted user. Indexes support these foreign-key operations. These fields are
record metadata, not a tamper-proof audit trail or full edit history.

Other metadata considered: soft-deletion state belongs to issue #63; publication
and import-source dates already belong to bibliographic fields. An edit history,
deletion actor, and optimistic-lock version are deferred until their application
behavior is defined.

## Tests

Run `make db-init-smoke` from the repository root. It initializes a fresh database
and executes `db/test/*_test.sql` with SQL errors treated as failures, including
the metadata tests. The existing CI database job runs this target. Tests cover
timestamp defaults and updates, actor constraints and deletion, system writes,
and timezone-independent instants; test data is rolled back.
