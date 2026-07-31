# DB

## Run

From the repository root, use `docker compose up db` to start the database.

If changes to the init script(s) in `./init-scripts/` are made or to reset the database, use `docker compose rm --stop --force --volumes db` to remove only the database container and its volume, then use `docker compose up db` again to start it with a fresh state.