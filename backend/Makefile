GO ?= go
COMPOSE ?= docker compose

-include .env
export DATABASE_URL HTTP_ADDRESS CORS_ALLOWED_ORIGINS

.PHONY: help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-serve run

help:
	@printf '%s\n' 'generate       Generate Echo interface, models, embedded OpenAPI' 'check          Check generated code, Go tests and vet' 'up             Start PostgreSQL -> migrations -> API -> Swagger UI' 'down           Stop containers (preserve database volume)' 'run            Apply migrations and run API locally (DATABASE_URL required)' 'migration-test Test up/down/up in an isolated database (TEST_DATABASE_URL required)' 'docs           Generate GoDoc example in HTML, RTF and text' 'docs-serve     Generate and view docs on localhost:6060'

generate:
	$(GO) tool oapi-codegen -config pkg/api/v1/oapi-codegen.yaml pkg/api/v1/openapi.yaml

generate-check:
	@tmp=$$(mktemp); cp pkg/api/v1/api.gen.go $$tmp; \
	trap 'rm -f "$$tmp"' EXIT; \
	$(MAKE) generate && cmp pkg/api/v1/api.gen.go $$tmp

build:
	$(GO) build -o bin/api ./cmd/api
	$(GO) build -o bin/migrate ./cmd/migrate

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: generate-check test vet

fmt:
	gofmt -w cmd internal migrations pkg/api/v1/spec.go

# Sequential recipes also preserve dependency order under make -j.
up:
	$(COMPOSE) up -d --wait postgres
	$(COMPOSE) build api migrate
	$(COMPOSE) up -d --wait api swagger

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f api migrate

migrate-up:
	$(COMPOSE) run --rm --build migrate up

migrate-down:
	$(COMPOSE) run --rm --build migrate down

migrate-status:
	$(COMPOSE) run --rm --build migrate status

run:
	$(GO) run ./cmd/migrate up
	$(GO) run ./cmd/api

migration-test:
	$(GO) test -tags=integration -count=1 ./internal/migration

docs:
	$(GO) run ./cmd/docgen

docs-serve:
	$(GO) run ./cmd/docgen -serve 127.0.0.1:6060
