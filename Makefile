.PHONY: secrets

SECRETS_DIR := secrets
SECRETS_FILES := \
	db_app_ro_password.txt \
	db_app_rw_password.txt \
	db_super_password.txt

secrets:
	@mkdir -p $(SECRETS_DIR)
	@for file in $(SECRETS_FILES); do \
		path="$(SECRETS_DIR)/$$file"; \
		if [ ! -f "$$path" ]; then \
			openssl rand -base64 24 > "$$path"; \
			echo "Created $$path"; \
		fi; \
	done
