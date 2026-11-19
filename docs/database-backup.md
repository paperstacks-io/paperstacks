# PostgreSQL full backups

Start the database with `docker compose up -d db` and wait for initialization
to finish. From a shell with Make and Docker Compose available, run:

```sh
make db-backup-full
```

The target enables password-authenticated replication connections for the
existing PostgreSQL administrator on directly attached networks. It then
starts a temporary container from the existing database image and reads the
administrator password from the existing Docker secret.

`pg_basebackup` takes a full physical copy of the entire PostgreSQL cluster
while the application can continue reading and writing. It includes the WAL
required to recover that backup. PostgreSQL 18 also supports incremental
backups, which require additional configuration and combining dependent
backups before recovery. This target creates full backups.

Each invocation writes a UTC timestamp directory into the Docker volume
`paperstacks_db_backups`. Existing directories cause the command to fail,
preventing an existing backup from being overwritten. `pg_verifybackup`
checks the completed backup against its manifest. A failed verification
leaves the directory for inspection. Only a successful command ending with
`Verified full backup` should be considered a verified backup.

List the saved directories:

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
bash db/test-backup.sh
```

The script creates an isolated Compose project with temporary credentials and
its own backup volume. It invokes the real Make target twice, checks online
reads and writes, then stops the source and restores the first backup into a
separate volume. It verifies the original test value and disables scheduled
jobs in the restored database. Test containers, volumes and the temporary
restore image are removed on exit.

`BACKUP_VOLUME` is configurable in the Make target so the test can use its
own volume.
