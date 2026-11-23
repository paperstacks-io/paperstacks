# Database backups and recovery

Run the commands below from the repository root.
For a new installation, create credentials with `make secrets` and
start the database with `docker compose up -d db` before taking a backup.

## Run a backup

```sh
make db-backup-full
make db-backup-incremental
```

| Argument | Meaning |
| --- | --- |
| `db-backup-full` | Create a complete backup that can be restored independently. |
| `db-backup-incremental` | Back up changes since the latest successful backup in the selected directory. Requires an existing backup. |
| `BACKUP_DIR=/absolute/path` | Choose the host directory where backups are saved. Defaults to `./backups`; relative paths are resolved from the directory where Make runs. |

For example:

```sh
make db-backup-full BACKUP_DIR=/srv/backups/paperstacks/postgres
make db-backup-incremental BACKUP_DIR=/srv/backups/paperstacks/postgres
```

Use the same `BACKUP_DIR` for a full backup and its subsequent incremental
backups. The host directory must be writable by the backup container. The
database remains available for reads and writes during either kind of backup.

## Expected result: full backup

A successful backup creates files on the host system like in this example:

```text
backups/
├── 2026-11-23_T120000Z/
│   ├── base.tar.gz
│   ├── pg_wal.tar.gz
│   └── backup_manifest
└── latest -> 2026-11-23_T120000Z
```

The backup contains the entire PostgreSQL cluster, including all databases,
roles, and the recovery logs needed to restore it. It does not include files
stored outside PostgreSQL. It is a physical backup, not a SQL dump.

Each full backup is independent of earlier backups. Creating a new full backup
starts a new incremental chain. Existing backups are retained; `latest` identifies
the most recent successful backup of either kind.

## Expected result: incremental backup

An incremental backup creates another timestamped directory with the same archive
filenames and prints the same success message. It uses the backup identified by
`latest` as its reference, then becomes the new `latest` itself.

It stores changed data relative to that reference and may be smaller than a full
backup; the size depends on database activity. It cannot be restored on its own.
For example, restoring the last backup in this sequence requires all three:

```text
2026-11-23_T120000Z  full
2026-11-23_T130000Z  incremental
2026-11-23_T140000Z  incremental
```

Keep the original full backup and every dependent incremental backup. Record
which backup starts each chain; directory names do not indicate the backup type.
If no reference backup is available, or PostgreSQL reports that the required WAL
summaries are unavailable, create a new full backup.

## Failed backups

If backup creation or verification fails, no new backup files are written to the
host backup directory. Existing backups and `latest` remain unchanged. The
command exits with an error; read its output before retrying.

A failure while transferring an already verified backup to the host, such as a
full disk or interrupted container, can leave an incomplete timestamp directory.
A failure while updating `latest` can also leave a backup that is not referenced
by `latest`. Do not treat a run as successful unless it exits successfully and
prints `Verified backup`.

If a retry reports a stale lock, first confirm that no backup is running, then
remove `.backup-lock` from the selected host backup directory.

## Restore a backup

These instructions target Linux, PostgreSQL 18, and the project's default
configuration without additional tablespaces. Install the PostgreSQL 18 tools
`pg_verifybackup`, `pg_waldump`, and `pg_combinebackup` on the host. The restored
server must use PostgreSQL 18 with the required extensions, including `pg_cron`;
the project's database service supplies those extensions.

Detailed information about the restore process will be provided later.
