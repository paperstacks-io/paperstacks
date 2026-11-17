.PHONY: secrets

SECRETS_DIR := secrets
SECRETS_FILES := \
	db_app_ro_password.txt \
	db_app_rw_password.txt \
	db_super_password.txt \
	db_barman_password.txt

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

.PHONY: db-backup-setup db-backup-check db-backup-full
COMPOSE ?= docker compose

db-backup-setup: ## Configure and start Barman for the running database
	@test -f "$(SECRETS_DIR)/db_barman_password.txt" && test -s "$(SECRETS_DIR)/db_barman_password.txt" || { echo "Run make secrets before configuring Barman" >&2; exit 1; }
	@{ cat "$(SECRETS_DIR)/db_barman_password.txt"; printf '\n'; cat db/backup/setup-role.sql; } | \
		$(COMPOSE) exec -T db sh -eu -c '\
			IFS= read -r PAPERSTACKS_BACKUP_PASSWORD; \
			case "$$PAPERSTACKS_BACKUP_PASSWORD" in \
				""|*[!a-zA-Z0-9_-]*) echo "Invalid backup secret" >&2; exit 1 ;; \
			esac; \
			export PAPERSTACKS_BACKUP_PASSWORD; \
			export PGPASSWORD="$$(cat /run/secrets/db_super_password)"; \
			psql -X -v ON_ERROR_STOP=1 -U postgres -d postgres; \
			hba_file="$$(psql -XAt -v ON_ERROR_STOP=1 -U postgres -d postgres -c "SHOW hba_file")"; \
			rule="host replication paperstacks_backup samenet scram-sha-256"; \
			if ! grep -Fxq "$$rule" "$$hba_file"; then \
				printf "\n%s\n" "$$rule" >> "$$hba_file"; \
			fi; \
			psql -X -v ON_ERROR_STOP=1 -U postgres -d postgres -c "SELECT pg_reload_conf()"'
	$(COMPOSE) --profile backup up -d --build --no-deps barman

db-backup-check: ## Check the Barman connection and WAL archive
	$(COMPOSE) run --rm --no-deps barman check paperstacks

db-backup-full: ## Create a full database backup with Barman
	$(COMPOSE) run --rm --no-deps barman backup --wait --wait-timeout 120 paperstacks

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

