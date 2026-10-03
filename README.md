# TourneyHub

Веб-платформа для командных и индивидуальных турниров.

- [`frontend/`](frontend/) — клиентское React-приложение.
- [`backend/`](backend/README.md) — сервер на Go и Echo, OpenAPI-контракты, миграции goose и Docker Compose с PostgreSQL.

## Запуск backend

Нужны Docker Compose и Make. Корневой Makefile передаёт команды в `backend/`; флаг `-C` указывать не требуется. Из корня репозитория:

```sh
cp backend/.env.example backend/.env
make up
```

Команда запускает PostgreSQL, применяет миграции и поднимает API со Swagger UI. По умолчанию API доступен на http://localhost:8080, Swagger UI — на http://localhost:8081. Настройки сервиса находятся в `backend/config/config.yaml`, логин и пароль БД — в переменных `DB_USER` и `DB_PASSWORD` из локального `backend/.env`. Порты контейнеров также задаются в `.env`.

Сейчас готовы контракты и инфраструктура; бизнес-операции возвращают `501 not_implemented`.

```sh
make generate
make check build docs
make down
```

Подробности архитектуры, миграций и примера GoDoc — в [документации backend](backend/README.md).
