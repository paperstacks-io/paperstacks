# DB

## Run

Use `docker compose up` to start the database.

If changes to the init script(s) in `./init-scripts/` are made or to reset the database, use `docker compose down -v` to stop the database and remove the volumes, then use `docker compose up` again to start it with a fresh state.


## Backend connection

Set `DATABASE_URL` to persist papers, users, and stacks in PostgreSQL. If unset, the backend uses its in-memory repositories. Use the `app_rw` role and its secret; do not commit the password:

```sh
export DATABASE_URL='postgres://app_rw:<URL-encoded-password>@localhost:5432/paperstacks?sslmode=disable'
```

Stack operations share the same connection pool as papers and users. Stack DAO
structs map database rows to the existing domain model. Stack creation and updates
write paper associations in the same transaction; association changes are
idempotent. Referenced users and papers must already exist. Deleting a stack
deletes its associations through the foreign key, leaving the papers intact.

Stack repository unit tests run with `make test` in `backend/` and require no
database. To additionally exercise the real schema, initialize a separate test
database and set `STACK_TEST_DATABASE_URL` to its `app_rw` connection string, then
run `make test-integration` in `backend/`. Without this variable only the PostgreSQL
test is skipped; the existing HTTP integration tests still run.
