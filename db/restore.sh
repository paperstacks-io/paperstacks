#!/bin/sh
set -eu
set -f
umask 077

if [ -z "${RESTORE_BACKUPS:-}" ]; then
    echo "Set BACKUP to a full backup, followed by any required incremental backups" >&2
    exit 1
fi

# Backup names are directory names, never paths or shell expressions.
for backup in $RESTORE_BACKUPS; do
    case "$backup" in
        latest|*partial|*[!a-zA-Z0-9_-]*)
            echo "Use explicit backup directory names: $backup" >&2
            exit 1
            ;;
    esac
    if [ ! -f "/backups/$backup/backup_manifest" ]; then
        echo "Missing backup manifest: $backup" >&2
        exit 1
    fi
done

if [ -n "$(ls -A /restore)" ]; then
    echo "Restore volume must be empty; use a new RESTORE_VOLUME" >&2
    exit 1
fi

# This also prevents concurrent restore commands from using the same volume.
mkdir /restore/.restore-work
mkdir /restore/.restore-work/backups
set --
for backup in $RESTORE_BACKUPS; do
    source="/backups/$backup"
    destination="/restore/.restore-work/backups/$backup"
    mkdir "$destination"

    if [ -f "$source/base.tar.gz" ]; then
        # The current backup setup uses only the default tablespaces.
        set +f
        for archive in "$source/"*.tar.gz; do
            case "$(basename "$archive")" in
                base.tar.gz|pg_wal.tar.gz) ;;
                *)
                    echo "Additional tablespaces require explicit restore mappings" >&2
                    exit 1
                    ;;
            esac
        done
        set -f
        pg_verifybackup --no-parse-wal "$source"
        tar -xzf "$source/base.tar.gz" -C "$destination"
        mkdir -p "$destination/pg_wal"
        tar -xzf "$source/pg_wal.tar.gz" -C "$destination/pg_wal"
        cp "$source/backup_manifest" "$destination/"
    else
        # Support plain-format backups created before gzip archives were added.
        cp -a "$source/." "$destination/"
    fi

    pg_verifybackup "$destination"
    set -- "$@" "$destination"
done

# pg_combinebackup validates ordering and dependencies, even for a single full.
pg_combinebackup -o /restore/.restore-work/data "$@"
pg_verifybackup /restore/.restore-work/data
chown -R postgres:postgres /restore/.restore-work/data
chmod 700 /restore/.restore-work/data
mv /restore/.restore-work/data /restore/data
rm -rf /restore/.restore-work

echo "Restore prepared and verified in /restore/data"
