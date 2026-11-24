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

COMPOSE_CMD := $(shell \
	if command -v docker >/dev/null 2>&1; then \
		echo "docker compose"; \
	elif command -v podman >/dev/null 2>&1; then \
		echo "podman compose"; \
	fi \
)

BACKUP_DIR ?= ./backups
.PHONY: db-backup-full db-backup-incremental
db-backup-full: ## Create a versioned full PostgreSQL backup in a host directory
db-backup-incremental: ## Back up changes since the latest successful backup

db-backup-full db-backup-incremental:
	BACKUP_MODE=$(patsubst db-backup-%,%,$@); \
	$(COMPOSE_CMD) run --rm --no-deps -T --entrypoint /bin/sh \
		-v "$(abspath $(BACKUP_DIR)):/backups" \
		-v "./db/backup.sh:/backup.sh:ro" \
		db /backup.sh $$BACKUP_MODE

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

