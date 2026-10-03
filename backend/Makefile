GO ?= go
COMPOSE ?= docker compose
PANDOC ?= pandoc

-include .env
export CONFIG_PATH DB_USER DB_PASSWORD DB_HOST DB_PORT HTTP_ADDRESS CORS_ALLOWED_ORIGINS

.PHONY: help generate generate-check build test vet check fmt up down logs migrate-up migrate-down migrate-status migration-test docs docs-export docs-serve run

help:
	@printf '%s\n' 'generate       Генерация моделей и HTTP-интерфейса Echo из OpenAPI' 'check          Проверка генерации, тесты Go и vet' 'up             Запуск PostgreSQL → миграций → API → Swagger UI' 'down           Остановка контейнеров с сохранением БД' 'run            Миграции и запуск API локально (DB_USER и DB_PASSWORD обязательны)' 'migration-test Проверка миграций в отдельной БД (TEST_DATABASE_URL обязателен)' 'docs           Документация модели встроенной командой go doc' 'docs-export    Экспорт документации в HTML и RTF через Pandoc' 'docs-serve     Просмотр HTML на localhost:6060 (нужен Python 3)'

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
	gofmt -w cmd internal pkg/api/v1/spec.go

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
