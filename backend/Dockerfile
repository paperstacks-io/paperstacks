# Build stage
ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS build

ARG APP_VERSION
ARG APP_GIT_HASH

RUN apk add --no-cache make

WORKDIR /app

COPY Makefile .
COPY go.mod .
COPY go.sum .
COPY cmd/ cmd/
COPY internal/ internal/

RUN make build-server

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/server .

EXPOSE 8080
ENV HOST 0.0.0.0

CMD ["./server"]
