# AGENTS.md

## Task Completion Requirements

- All of `make build`, `make test`, and `make test-integration` must pass before considering tasks completed.

## Project Snapshot

Paperstacks is a platform to organize and share academic literature.

This repository is a VERY EARLY WIP. Proposing sweeping changes that improve long-term maintainability is encouraged.

## Core Priorities

1. Performance first.
2. Reliability first.
3. Keep behavior predictable under load and during failures (session restarts, reconnects, partial streams).

If a tradeoff is required, choose correctness and robustness over short-term convenience.

## Maintainability

Long term maintainability is a core priority. Domain-driven design is encouraged. Don't be afraid to change existing code. Don't take shortcuts by just adding local logic to solve a problem.

## Package role

- `internal/common`: Shared packages.
- `internal/web`: Web UI frontend using HTMX and http/template.
- `internal/server`: HTTP server to serve API.
- `internal/paper`: All packages relevant to the domain Paper.
- `internal/stack`: All packages relevant to the domain Stack.
- `internal/doi`: All packages relevant to the domain DOI.
- `internal/integration`: Integration tests that run against the HTTP API.

## Instructions for Claude/Codex

- Always plan before you start coding.
