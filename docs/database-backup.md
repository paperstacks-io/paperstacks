# PostgreSQL backups with Barman

Barman is opt-in via the `backup` profile in `compose.yaml`. It runs in its
own container alongside PostgreSQL and stores backups and streamed
WAL in the separate `barman_data` volume. A failure of the database container
does not remove this volume. A host or disk failure can affect both volumes;
keep a copy outside the database host before relying on this setup for recovery.

## Start and check

From the project root in a Bash environment:

```bash
make secrets
docker compose up -d db
# Wait for PostgreSQL to finish initialization.
make db-backup-setup
```

Setup creates or updates the dedicated `paperstacks_backup` role on the
running database, adds a password-authenticated replication rule for this role
on the database container's directly attached networks, reloads PostgreSQL's
configuration and starts the Barman service. It works with an already
initialized cluster. The role can stream the entire cluster and read
monitoring information, but is not a superuser. Keep its secret and the
backup repository private. Only the setup command uses the PostgreSQL
administrator account.

The service runs Barman maintenance every 30 seconds to maintain the physical
replication slot, keep receiving WAL, and process completed segments.
Setup also registers the `barman-weekly-full` job in `pg_cron`, which requests
a full backup every Sunday at 02:00 UTC. The first setup requests an initial
backup. Run setup again to install the job on an existing deployment.

Barman checks for pending requests every minute. It processes one full backup
at a time while WAL maintenance continues independently. The request remains
pending until the backup command completes with the required WAL archived.
If Barman is temporarily unavailable, outstanding requests are picked up
when it returns and collapsed into one backup of the current cluster.

Check the backup service or trigger a manual backup if needed:

```bash
docker compose run --rm --no-deps barman \
  switch-wal --archive --archive-timeout 120 paperstacks
make db-backup-check
make db-backup-full
docker compose run --rm --no-deps barman \
  list-backups paperstacks
```

Check that the new backup is `DONE`, not `WAITING_FOR_WALS`. The backup can
run while the app reads and writes. It contains the entire PostgreSQL cluster;
PDFs stored outside PostgreSQL need a separate backup.

## Scope

Barman supports incremental backups. The configured schedule uses weekly full
backups with continuous WAL archiving. The schedule does not set a retention
policy. Do not use a Compose
`down -v` command against a repository you want to retain.

If Barman stops receiving WAL, its replication slot can retain WAL on the
database host. Monitor the receiver and free disk space. Before permanently
decommissioning Barman, plan removal of the `paperstacks_barman` slot; merely
stopping the container leaves the slot in place.

Sources: [Barman WAL archiving](https://docs.pgbarman.org/release/3.20.1/user_guide/wal_archiving.html),
[Barman backup commands](https://docs.pgbarman.org/release/3.20.1/user_guide/commands.html),
[PostgreSQL replication slots](https://www.postgresql.org/docs/18/warm-standby.html#STREAMING-REPLICATION-SLOTS).
