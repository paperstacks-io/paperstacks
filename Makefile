.PHONY: secrets

SECRETS_DIR := secrets
SECRETS_FILES := \
	db_app_ro_password.txt \
	db_app_rw_password.txt \
	db_super_password.txt

secrets: ## Create secret credential files if they don't exist
	@mkdir -p $(SECRETS_DIR)
	@for file in $(SECRETS_FILES); do \
		path="$(SECRETS_DIR)/$$file"; \
		if [ ! -f "$$path" ]; then \
			openssl rand -base64 32 | tr '+/' '-_' | tr -d '=\n' > "$$path"; \
			echo "Created $$path"; \
		fi; \
	done

.PHONY: db-init-smoke
db-init-smoke: ## Smoke test Postgres init scripts against a fresh database
	./db/smoke-init.sh

.PHONY: help
.PHONY: db-backup-full
BACKUP_VOLUME ?= paperstacks_db_backups
db-backup-full: ## Create a versioned full PostgreSQL backup in a Docker volume
	docker compose exec -T --user postgres db sh -eu -c '\
		rule="host replication postgres samenet scram-sha-256"; \
		if ! grep -Fxq "$$rule" "$$PGDATA/pg_hba.conf"; then \
			printf "\n%s\n" "$$rule" >> "$$PGDATA/pg_hba.conf"; \
		fi; \
		psql -X -v ON_ERROR_STOP=1 -U postgres -d postgres -c "SELECT pg_reload_conf()"'
	docker compose run --rm --no-deps -T --entrypoint /bin/sh \
		-v "$(BACKUP_VOLUME):/backups" db -eu -c '\
		export PGPASSWORD="$$(cat /run/secrets/db_super_password)"; \
		backup_dir="/backups/$$(date -u +%Y%m%dT%H%M%SZ)"; \
		mkdir "$$backup_dir"; \
		pg_basebackup --host=db --username=postgres --no-password \
			--pgdata="$$backup_dir" --wal-method=stream --checkpoint=fast --progress; \
		pg_verifybackup "$$backup_dir"; \
		echo "Verified full backup: $$backup_dir"'

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

