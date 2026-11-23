# PostgreSQL full backups

Create the credential files with `make secrets`, then start the database with
`docker compose up -d db` and wait for initialization
to finish. From a shell with Make and Docker Compose available, run:

```sh
make db-backup-full
```

The database init scripts create the dedicated `paperstacks_backup` login
with replication privileges and enable password-authenticated replication
connections on directly attached networks. The backup target starts a
temporary container from the existing database image and reads the backup
password from the `db_backup_password` Docker secret.

The Make targets run the shared `db/backup.sh` script inside the temporary
container. The script selects full or incremental mode, creates and verifies
the archives, and updates `latest`. The container is removed after the command.

For an existing database, run `make secrets` and recreate the database
container with `docker compose up -d db` to mount the new secret. Then apply
the role initialization and access setup once. Re-running `001_init.sql`
also reapplies the application role passwords and permissions from the current
secrets and initialization definitions:

```sh
docker compose exec -T --user postgres db psql -X -v ON_ERROR_STOP=1 \
  -U postgres -d postgres -f /docker-entrypoint-initdb.d/001_init.sql
docker compose exec -T --user postgres db \
  sh /docker-entrypoint-initdb.d/008_configure_backup_access.sh
```

`pg_basebackup` takes a full physical copy of the entire PostgreSQL cluster
while the application can continue reading and writing. It includes the WAL
required to recover that backup. For subsequent changes, run:

```sh
make db-backup-incremental
```

This uses the manifest referenced by `latest`, which points to the last
successfully verified full or incremental backup in the same volume.
Create a new full backup first when adopting this mechanism for an older
backup volume without a `latest` link. A new full backup starts a new chain.
PostgreSQL must have `summarize_wal=on`, as configured in Compose. Apply the
updated Compose configuration to existing database containers before taking
the new base backup. If required WAL summaries have expired, take a new full
backup.

Both commands store gzip-compressed archives. A backup is first written to
a `.partial` directory and published under its timestamp after verification.
Failed runs leave partial directories for inspection and keep `latest`
pointing to the previous backup. A volume-local lock prevents overlapping
commands. After an abrupt container termination, remove `.backup-lock`
only after confirming that no backup is still running.

For incremental recovery, extract the full backup and every dependent
incremental backup into separate directories, including each WAL archive
into its respective `pg_wal` directory and its manifest into the root.
Verify each extracted backup with `pg_verifybackup`, then combine the chain
in chronological order:

```sh
pg_combinebackup -o /restore /extracted/full /extracted/incremental-1 /extracted/incremental-2
pg_verifybackup /restore
```

Start PostgreSQL using the combined directory. Keep the complete dependency
chain for every incremental backup you want to recover. PostgreSQL checks
the manifest against the running cluster and rejects incompatible bases.

Each invocation writes a UTC timestamp directory into the Docker volume
`paperstacks_db_backups`. Existing directories cause the command to fail,
preventing an existing backup from being overwritten. `pg_verifybackup`
checks the completed backup against its manifest using `--no-parse-wal`.
PostgreSQL 18 cannot parse WAL directly inside tar archives. The function
test additionally verifies the extracted backup, including its required WAL.
A failed verification
leaves the directory for inspection. Only a successful command ending with
`Verified backup` should be considered a verified backup.

List the saved directories:

Each backup uses gzip-compressed tar format. With the current default
tablespace configuration, each timestamp directory contains three files:
`base.tar.gz`, `pg_wal.tar.gz` and `backup_manifest`. Additional tablespaces
produce additional archives. Extract `base.tar.gz` into the restore data
directory and `pg_wal.tar.gz` into its `pg_wal` subdirectory before starting
PostgreSQL.

```sh
docker compose run --rm --no-deps -T --entrypoint /bin/sh \
  -v paperstacks_db_backups:/backups:ro db -c 'ls -1 /backups'
```

The volume contains administrator credentials and database contents, and
shares the Docker host's storage. Copy backups to separate storage for
protection against host or disk failure. Backup retention is manual.
Files stored outside PostgreSQL need their own backup.

The backup contains the source cluster's `pg_cron` extension. A restore test
must use a PostgreSQL image with that extension installed and keep scheduled
jobs disabled in the isolated restored instance.

See the [PostgreSQL pg_basebackup documentation](https://www.postgresql.org/docs/18/app-pgbasebackup.html).

## Function test

With Bash, Make and Docker Compose available, run from the repository root:

```sh
bash db/test-restore.sh
```

The script creates an isolated Compose project with temporary credentials and
its own backup volume. It invokes the real Make targets for a full backup and
two chained incremental backups, then verifies both restored states through
the Compose restore service. It also checks rejected inputs and that
scheduled jobs are disabled. Test containers and volumes are removed on exit.


```sh
bash db/test-backup.sh
```

The script creates an isolated Compose project with temporary credentials and
its own backup volume. It invokes the real Make targets for two full backups
and two chained incremental backups, checks online reads and writes, then
stops the source. It restores the first full backup and separately combines
and restores the incremental chain. It verifies the expected test values and disables scheduled
jobs in the restored database. Test containers, volumes and the temporary
restore image are removed on exit.

`BACKUP_VOLUME` is configurable in the Make target so the test can use its
own volume.
