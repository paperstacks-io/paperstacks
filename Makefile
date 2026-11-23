.PHONY: secrets

SECRETS_DIR := secrets
SECRETS_FILES := \
	db_app_ro_password.txt \
	db_app_rw_password.txt \
	db_super_password.txt \
	db_backup_password.txt

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

BACKUP_VOLUME ?= paperstacks_db_backups
.PHONY: db-backup-full db-backup-incremental
db-backup-full: ## Create a versioned full PostgreSQL backup in a Docker volume
db-backup-incremental: ## Back up changes since the latest successful backup

db-backup-full db-backup-incremental:
	docker compose run --rm --no-deps -T --entrypoint /bin/sh \
		-v "$(BACKUP_VOLUME):/backups" \
		-v "./db/backup.sh:/db/backup.sh:ro" \
		db /db/backup.sh $(patsubst db-backup-%,%,$@)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

