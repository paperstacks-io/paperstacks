# DB

## Run

Use `docker compose up` to start the database.

If changes to the init script(s) in `./init-scripts/` are made or to reset the database, use `docker compose down -v` to stop the database and remove the volumes, then use `docker compose up` again to start it with a fresh state.