# Restore PostgreSQL backups

The restore command prepares a verified PostgreSQL 18 data directory in a
separate Docker volume. It reads the backup volume without modifying it.
The existing database can remain online during preparation.

Run `make secrets` before using Compose. Use a trusted backup from the same
PostgreSQL major version and compatible extension versions. The Compose
database entrypoint installs the existing `pg_cron` extension when starting
the restored database.

## Select and prepare a backup

List backups in the default backup volume:

```sh
docker compose run --rm --no-deps -T --entrypoint sh \
  -v paperstacks_db_backups:/backups:ro db -c 'ls -1 /backups'
```

Use the explicit timestamp directory name. Restore a full backup with:

```sh
make db-restore BACKUP=20261123T120000Z RESTORE_VOLUME=paperstacks_restore_full
```

For an incremental backup, provide its full dependency chain from oldest to
newest. For example, a full backup followed by two dependent increments:

```sh
make db-restore BACKUP="20261123T120000Z 20261124T120000Z 20261125T120000Z" \
  RESTORE_VOLUME=paperstacks_restore_incremental
```

Set `BACKUP_VOLUME` to select a different backup volume. The default restore
volume is `paperstacks_db_restore`. It must be empty. A failed attempt can
leave temporary data there; choose a new volume for the next attempt.

The script supports gzip-compressed tar backups and older plain-format
backups using the default tablespaces. It checks manifests, extracts each
backup, verifies the required WAL, and runs `pg_combinebackup` to validate
the chain and produce the restore directory. It verifies the result and sets
ownership and permissions for PostgreSQL. Additional tablespaces require
explicit restore mappings and are rejected by the archive restore.

Success is reported as `Restore prepared and verified in /restore/data`.
This prepares the data; server startup and application checks follow.

## Review and switch the database

The `db-restored` service is part of `compose.yaml` and enabled through the
`restore` profile. It can run alongside the original database for inspection.
It has its own Docker hostname and does not publish a host port.

```sh
export RESTORE_VOLUME=paperstacks_restore_full
docker compose --profile restore up -d --no-deps db-restored
docker compose logs --tail 50 db-restored
docker compose exec -T --user postgres db-restored \
  psql -X -v ON_ERROR_STOP=1 -U postgres -d paperstacks -c 'SELECT count(*) FROM public.paper'
```

Check expected records and credentials. After accepting the recovered state,
stop the app and the original database and point the app at `db-restored`.
Replace the password placeholder with the restored application role's
URL-encoded password:

```sh
docker compose stop paperstacks db
export DATABASE_URL='postgresql://app_rw:<URL-encoded-password>@db-restored:5432/paperstacks'
docker compose up -d --no-deps paperstacks
```

Persist `RESTORE_VOLUME` and `DATABASE_URL` in the deployment environment.
Use the `restore` profile to include `db-restored` in subsequent deployments.
Keep the original database volume and source backups until recovery has been
accepted.

`db-restored` disables scheduled `pg_cron` jobs so restored jobs
cannot immediately modify recovered data. Review jobs before deliberately
enabling them in that deployment's configuration.

Roles and passwords are restored from the backup. Application credentials
must match those restored roles, or be rotated before the application
reconnects. Restore any PDFs or other files stored outside PostgreSQL
separately.

For subsequent backups of the restored database, select its hostname and
start a new full-backup chain:

```sh
make db-backup-full BACKUP_HOST=db-restored
make db-backup-incremental BACKUP_HOST=db-restored
```

## Isolated function test

With Bash, Make and Docker Compose available, run:

```sh
bash db/test-restore.sh
```

The test creates its own source database and backup volume, takes a full
backup and two increments, and uses the actual Make restore command. It
rejects missing backups, nonempty targets and an incomplete chain. It starts
both restored states through the `db-restored` service, verifies their
data and confirms that scheduled jobs are disabled. Test resources are
removed on exit.

Sources are the PostgreSQL documentation for
[pg_combinebackup](https://www.postgresql.org/docs/18/app-pgcombinebackup.html)
and [pg_verifybackup](https://www.postgresql.org/docs/18/app-pgverifybackup.html).
