# Build stage
ARG GO_VERSION=tip
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache make

WORKDIR /app

COPY Makefile .
COPY go.mod .
COPY cmd/ cmd/
COPY internal/ internal/

RUN make build-server

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=build /app/server .

EXPOSE 8080

CMD ["./server"]
