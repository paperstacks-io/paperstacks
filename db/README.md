# DB

## Run

Use `docker compose up` to start the database.

If changes to the init script(s) in `./init-scripts/` are made or to reset the database, use `docker compose down -v` to stop the database and remove the volumes, then use `docker compose up` again to start it with a fresh state.


## Backend connection

Set `DATABASE_URL` to persist papers in PostgreSQL. If unset, the backend uses its in-memory repository. Use the `app_rw` role and its secret; do not commit the password:

```sh
export DATABASE_URL='postgres://app_rw:<URL-encoded-password>@localhost:5432/paperstacks?sslmode=disable'
```