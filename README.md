# TourneyHub

Веб-платформа для командных и индивидуальных турниров.

- [`frontend/`](frontend/) — клиентское React-приложение.
- [`backend/`](backend/README.md) — сервер на Go и Echo, OpenAPI-контракты, миграции goose и Docker Compose с PostgreSQL.

## Запуск backend

Нужны Docker Compose и Make. Из корня репозитория:

```sh
cp backend/.env.example backend/.env
make -C backend up
```

Команда запускает PostgreSQL, применяет миграции и поднимает API со Swagger UI. По умолчанию API доступен на http://localhost:8080, Swagger UI — на http://localhost:8081. Порты настраиваются в `backend/.env`.

Сейчас готовы контракты и инфраструктура; бизнес-операции возвращают `501 not_implemented`.

```sh
make -C backend generate
make -C backend check build docs
make -C backend down
```

Подробности архитектуры, миграций и примера GoDoc — в [документации backend](backend/README.md).
