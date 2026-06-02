# Paperstacks Backend

Backend for the Paperstacks application, written in Go.

## Architecture Overview

![architecture overview](../docs/architecture-backend.svg)

## Prerequisites

- Go 1.26 or later

Install dependencies:

```bash
make deps
```

## Building

To build the server binary:

```bash
make build
```

This will compile the application and create a `server` executable in the current directory.

To clean up build artifacts:

```bash
make clean
```

## Running

Start the server with:

```bash
./server
```

The server will start on `localhost:8080` by default.

### Configuration

You can customize the server behavior with environment variables:

- `HOST` - Server host (default: `localhost`)
- `PORT` - Server port (default: `8080`)

Example:

```bash
HOST=0.0.0.0 PORT=9000 ./server
```

## Testing

Run all unit tests:

```bash
make test
```

Run integration tests:

```bash
make test-integration
```

Test coverage is generated in `coverage.out`.

## Development

Run the server with live reload:

```bash
air
```

If you make modifications on the HTMX/CSS part also run:

```bash
npm run watch:css
```

## HTTP API

`openapi.yaml` contains a API documentation that can be viewed with
[Swagger Editor](https://editor.swagger.io/?url=https://raw.githubusercontent.com/paperstacks-io/paperstacks/main/backend/openapi.yaml)
or [Redoc](https://redocly.github.io/redoc/?url=https://raw.githubusercontent.com/paperstacks-io/paperstacks/main/backend/openapi.yaml).
