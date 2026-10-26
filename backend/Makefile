GO ?= go
GENERATION_TOOLCHAIN := go1.25.8
COMPOSE ?= docker compose
PANDOC ?= pandoc

-include .env
export CONFIG_PATH DB_USER DB_PASSWORD DB_HOST DB_PORT HTTP_ADDRESS CORS_ALLOWED_ORIGINS
export PUBLIC_WEB_URL SMTP_ADDRESS SMTP_FROM SMTP_USER SMTP_PASSWORD SMTP_ALLOW_INSECURE SMTP_DIAL_TIMEOUT SMTP_TIMEOUT
export WORKER_INTERVAL EVENT_RETENTION HTTP_READ_HEADER_TIMEOUT HTTP_READ_TIMEOUT HTTP_IDLE_TIMEOUT HTTP_SHUTDOWN_TIMEOUT HTTP_BODY_LIMIT SSE_POLL_INTERVAL SSE_HEARTBEAT_INTERVAL
export ADMIN_EMAIL ADMIN_NICKNAME ADMIN_PASSWORD

.PHONY: help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run

help:
	@printf '%s\n' 'generate       Генерация моделей и HTTP-интерфейса Echo из OpenAPI' 'check          Проверка генерации, тесты Go и vet' 'up             Запуск PostgreSQL → миграций → API → Swagger UI' 'down           Остановка контейнеров с сохранением БД' 'run            Миграции и запуск API локально (DB_USER и DB_PASSWORD обязательны)' 'migration-test Проверка миграций в отдельной БД (TEST_DATABASE_URL обязателен)' 'docs           Документация модели встроенной командой go doc' 'docs-export    Экспорт документации в HTML и RTF через Pandoc' 'docs-serve     Просмотр HTML на localhost:6060 (нужен Python 3)'

generate:
	GOTOOLCHAIN=$(GENERATION_TOOLCHAIN) $(GO) tool oapi-codegen -config pkg/api/v1/oapi-codegen.yaml pkg/api/v1/openapi.yaml
	GOTOOLCHAIN=$(GENERATION_TOOLCHAIN) $(GO) run ./tools/dtomap

generate-check:
	@set -e; tmp=$$(mktemp -d); \
	cp pkg/api/v1/api.gen.go "$$tmp/api.gen.go"; \
	cp internal/api/v1/mapping.gen.go "$$tmp/mapping.gen.go"; \
	trap 'rm -rf "$$tmp"' EXIT; \
	$(MAKE) generate; \
	cmp pkg/api/v1/api.gen.go "$$tmp/api.gen.go"; \
	cmp internal/api/v1/mapping.gen.go "$$tmp/mapping.gen.go"

build:
	$(GO) build -o bin/api ./cmd/api
	$(GO) build -o bin/migrate ./cmd/migrate
	$(GO) build -o bin/bootstrap ./cmd/bootstrap

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: generate-check test vet

fmt:
	gofmt -w cmd internal tools pkg/api/v1/spec.go

# Последовательные команды сохраняют порядок запуска даже при make -j.
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

.PHONY: integration-test metrics bootstrap
integration-test:
	$(GO) test -tags=integration -race -count=1 ./...

metrics:
	mkdir -p build/metrics
	$(GO) run ./cmd/metrics -out build/metrics/current.json
	$(GO) tool gocyclo -over 10 -ignore 'api.gen.go|_test.go|cmd/metrics/' cmd internal pkg > build/metrics/gocyclo.txt || test $$? -eq 1
	$(GO) tool gocognit -over 15 -ignore 'api.gen.go|_test.go|cmd/metrics/' cmd internal pkg > build/metrics/gocognit.txt || test $$? -eq 1

bootstrap:
	$(GO) run ./cmd/bootstrap

docs:
	mkdir -p build/docs
	$(GO) doc -all ./internal/model > build/docs/example.txt

docs-export: docs
	@command -v $(PANDOC) >/dev/null || { printf '%s\n' 'Для экспорта HTML и RTF установите Pandoc.'; exit 1; }
	@{ printf '%s\n' '# TourneyHub — доменная модель' '' '[Исходный файл rating.go](_source/rating.go) · [Скачать RTF](example.rtf)' '' '```text'; cat build/docs/example.txt; printf '%s\n' '```'; } > build/docs/example.md
	mkdir -p build/docs/_source
	cp internal/model/rating.go build/docs/_source/rating.go
	$(PANDOC) build/docs/example.md --standalone --metadata lang=ru --metadata pagetitle='TourneyHub — доменная модель' --to=html --output=build/docs/index.html
	$(PANDOC) build/docs/example.md --standalone --to=rtf --output=build/docs/example.rtf

docs-serve: docs-export
	python3 -m http.server 6060 --bind 127.0.0.1 --directory build/docs
