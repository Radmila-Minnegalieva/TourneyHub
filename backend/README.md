# Запуск TourneyHub Backend

TourneyHub — платформа для организации и проведения турниров. Backend предоставляет REST API для аккаунтов, команд, турниров, заявок, жеребьёвки, расписания, результатов, споров, рейтингов и уведомлений. Фактический стек: Go 1.25+, Echo 4, pgx 5, PostgreSQL 18.6 и goose. HTTP-контракт: [OpenAPI 3.0.3](pkg/api/v1/openapi.yaml).

Основной способ локального запуска — Docker Compose. Он запускает PostgreSQL, применяет миграции и поднимает API, Swagger UI и Mailpit для просмотра писем. Frontend запускается отдельно. Встроенного демонстрационного наполнения дисциплин и турниров нет: данные создаются через API.

## 1 Получить исходники и подготовить инструменты

Для запуска целиком в контейнерах нужны Git, работающий Docker Desktop либо Docker Engine с **Compose v2**, Make и доступ к сети для загрузки образов. Для примеров запросов нужен curl; их можно выполнить и через Swagger UI. На Windows команды удобнее выполнять в WSL2 с включённой интеграцией Docker Desktop.

```sh
git clone --branch backend/lab3-complete git@github.com:Radmila-Minnegalieva/TourneyHub.git
cd TourneyHub/backend
docker --version
docker compose version
make --version
```

Команда использует SSH-доступ к GitHub. При подключении через HTTPS замените URL на `https://github.com/Radmila-Minnegalieva/TourneyHub.git`, сохранив `--branch backend/lab3-complete`. После слияния pull request можно клонировать обычную `main` без параметра `--branch`.

Если репозиторий уже есть, сохраните свою работу, получите ветку backend и откройте каталог `backend/`. До слияния pull request используйте `backend/lab3-complete`, а не прежнюю версию `main`.

Все дальнейшие команды выполняются **из `TourneyHub/backend/`**, если явно не указан другой каталог. Команды Make можно вызывать и из корня TourneyHub: корневой Makefile перенаправляет их в backend. Для самостоятельных команд `docker compose` каталог backend обязателен.

Для запуска Go-процесса на хосте, тестов, генерации и метрик дополнительно установите Go **1.25 или новее**. В контейнере компилятор уже есть: Dockerfile использует Go 1.25.8, поэтому Docker-вариант не требует локальной установки Go.

## 2 Настроить локальное окружение

Создайте `.env`, только если его ещё нет:

```sh
if [ ! -f .env ]; then
  cp .env.example .env
fi
```

Откройте `.env` в редакторе и проверьте значения:

```dotenv
DB_USER=tourneyhub
DB_PASSWORD=tourneyhub_local
POSTGRES_PORT=5433
API_PORT=8080
SWAGGER_PORT=8081
MAILPIT_PORT=8025
PUBLIC_WEB_URL=http://localhost:5173
```

Это **демонстрационные локальные значения** из `.env.example`. Для собственного стенда выберите свои реквизиты. `.env` исключён из Git; не заменяйте уже настроенный файл повторным копированием примера.

`DB_USER` и `DB_PASSWORD` обязательны. При первом создании volume образ PostgreSQL использует их для инициализации роли. Изменение этих строк после создания БД **не меняет пароль или пользователя в существующем PostgreSQL**. Для существующих данных используйте действующие реквизиты либо отдельно измените роль в БД; удалять volume для исправления пароля не требуется.

## 3 Запустить контейнеры

```sh
make up
```

Команда последовательно запускает PostgreSQL, собирает образ приложения и запускает миграции, API и Swagger UI. API начинает работу после успешных миграций; Mailpit поднимается как его зависимость. Первый запуск может занять несколько минут из-за загрузки образов и модулей Go.

Проверка состояния и логов:

```sh
docker compose ps -a
docker compose logs --tail=100 migrate api
curl --fail http://localhost:8080/health/live
curl --fail http://localhost:8080/health/ready
```

Оба health-маршрута должны вернуть `{"status":"ok"}`. Проверка `/health/ready` обращается к БД. Контейнер `migrate` после успешного выполнения имеет состояние `Exited (0)` — он выполняет разовую задачу. Контейнер API должен быть `healthy`.

При значениях портов из примера доступны:

| Что открыть | Адрес |
| --- | --- |
| Корень REST API | http://localhost:8080/api/v1 |
| Swagger UI с операциями API | http://localhost:8081 |
| OpenAPI YAML | http://localhost:8080/openapi.yaml |
| Почтовый ящик Mailpit | http://localhost:8025 |
| PostgreSQL с хоста | localhost:5433, база `tourneyhub` |
| Проверка готовности API | http://localhost:8080/health/ready |

`/api/v1` — префикс маршрутов, а не отдельная веб-страница. Отсутствие HTML по этому адресу не означает сбой. Порты изменяются в `.env`; после изменения используйте соответствующие адреса и повторите `make up`. Swagger берёт адрес API из `API_PORT` автоматически.

## 4 Создать первого администратора

Обычная регистрация создаёт только участника. Для назначения организаторов и судей, создания дисциплин и доступа к администрированию нужен первый администратор. Создайте его после успешного запуска миграций.

В Bash или Zsh задайте адрес и nickname, затем введите пароль без отображения на экране:

```sh
export ADMIN_EMAIL='admin@example.test'
export ADMIN_NICKNAME='admin'
printf 'Пароль администратора: '
read -r -s ADMIN_PASSWORD
printf '\n'
export ADMIN_PASSWORD

docker compose run --rm --no-deps \
  -e ADMIN_EMAIL -e ADMIN_NICKNAME -e ADMIN_PASSWORD \
  api /app/bootstrap

unset ADMIN_PASSWORD
```

Пароль должен содержать не менее 12 символов и не более 72 байт UTF-8. Nickname — 3–32 латинские буквы, цифры или `_`. Адрес и nickname примера можно заменить своими. Параметры `-e` передают значения из текущего терминала во временный контейнер; пароль не нужно записывать в командную строку или репозиторий.

Успешный запуск выводит `administrator is ready`. Аккаунт уже подтверждён и получает роли `participant`, `organizer`, `referee`, `admin`. Повторный bootstrap того же действующего администратора не меняет его пароль. Если первый администратор уже существует, команда не создаст другого: дополнительные аккаунты регистрируются обычным способом, а роли назначаются через защищённый API действующим администратором.

Войдите через `POST /api/v1/auth/login`; в Swagger нажмите **Authorize** и вставьте `access_token`. После этого создайте нужные дисциплины через `POST /admin/disciplines` и назначьте роли другим пользователям. Поля запросов смотрите в Swagger; заранее заполненного справочника нет.

## 5 Проверить регистрацию и сессию

Следующий пример использует отдельный демонстрационный аккаунт. Если он уже создан, выберите другой e-mail и nickname. Значение пароля в примере предназначено только для локальной проверки.

```sh
export API_BASE_URL='http://localhost:8080/api/v1'

curl -i -X POST "$API_BASE_URL/auth/register" \
  -H 'Content-Type: application/json' \
  --data '{"email":"player@example.test","password":"DemoPassword_2026","nickname":"player_demo","personal_data_consent":true}'
```

Ожидается **202 Accepted**. Затем откройте Mailpit на http://localhost:8025. Фоновый обработчик проверяет почту при старте и далее раз в минуту; письмо после регистрации может прийти не сразу. Скопируйте значение `token` из ссылки подтверждения. Пока соответствующего frontend-экрана нет, ссылка на порт 5173 сама аккаунт не подтверждает: вызовите API вручную.

```sh
curl -i -X POST "$API_BASE_URL/auth/verify-email" \
  -H 'Content-Type: application/json' \
  --data '{"token":"ВСТАВЬТЕ_ТОКЕН_ИЗ_ПИСЬМА"}'

curl -i -X POST "$API_BASE_URL/auth/login" \
  -H 'Content-Type: application/json' \
  --data '{"email":"player@example.test","password":"DemoPassword_2026"}'
```

Подтверждение возвращает **204 No Content**, вход — **200 OK** с `access_token`, `refresh_token`, `token_type: "Bearer"` и `expires_in: 900`. Access действует 15 минут, refresh — 30 дней. В Swagger либо в терминале используйте access-токен для защищённого профиля:

```sh
printf 'Access token из ответа login: '
read -r ACCESS_TOKEN
curl -i "$API_BASE_URL/users/me" -H "Authorization: Bearer $ACCESS_TOKEN"
unset ACCESS_TOKEN
```

Для обновления пары отправьте `POST /auth/refresh` с телом `{"token":"REFRESH_TOKEN"}`. Сохраните **оба новых токена**: старый refresh после ротации повторно не используется. Выход — `POST /auth/logout` с таким же полем `token`, содержащим текущий refresh. Нельзя отправлять `refresh_token` вместо поля `token`: схема запроса называется `TokenRequest` и использует именно `token`.

Вход неподтверждённого пользователя отклоняется. После пяти попыток входа с одного IP за минуту можно получить 429; дождитесь нового окна. Пользовательский пароль также ограничен 72 байтами, что важно для кириллицы и других многобайтных символов.

## 6 Запуск Go на хосте

Этот вариант нужен для разработки с быстрым перезапуском Go-процесса. PostgreSQL остаётся в Docker. Если весь Compose уже запущен, сначала освободите порт API:

```sh
docker compose stop api
docker compose up -d --wait postgres
go version
go mod download
make run
```

`make run` экспортирует настройки из `.env`, применяет миграции и запускает API. YAML по умолчанию подключает Go-процесс к `localhost:5433`. Если в `.env` другой `POSTGRES_PORT`, укажите тот же порт через `make run DB_PORT=ВАШ_ПОРТ`; `POSTGRES_PORT` предназначен для Compose и сам не меняет `database.port` в Go.

Самостоятельный `go run ./cmd/api` **не читает `.env`**. Для него сначала экспортируйте `DB_USER`, `DB_PASSWORD` и нужные переопределения либо используйте Make. При запуске готового бинарника также передайте окружение и корректный `CONFIG_PATH`; относительный путь к YAML считается от рабочего каталога процесса.

Чтобы создать первого администратора при этом варианте, в другом терминале из `backend/` задайте `ADMIN_EMAIL`, `ADMIN_NICKNAME` и `ADMIN_PASSWORD`, как в разделе 4, затем выполните `make bootstrap` вместо `docker compose run ... /app/bootstrap`. Если порт БД менялся, передайте тот же `DB_PORT` и этой команде. После завершения удалите пароль из окружения через `unset ADMIN_PASSWORD`.

### Доставка писем при Go-процессе на хосте

В обычном Compose SMTP-порт Mailpit доступен только контейнерам как `mailpit:1025`. Наружу опубликован его веб-интерфейс, но не SMTP. Кроме того, в YAML для host-варианта `smtp.address` пуст: письма сохраняются в outbox до настройки SMTP.

Чтобы доставлять письма в Mailpit из локального Go, добавьте временное Compose-переопределение с SMTP-портом:

```sh
cat > /tmp/tourneyhub-host-mailpit.yaml <<'YAML'
services:
  mailpit:
    ports:
      - "127.0.0.1:1025:1025"
YAML

docker compose -f compose.yaml -f /tmp/tourneyhub-host-mailpit.yaml up -d postgres mailpit
make run SMTP_ADDRESS=localhost:1025 SMTP_ALLOW_INSECURE=true
```

Веб-интерфейс писем остаётся на порту `MAILPIT_PORT`, по умолчанию 8025. Если нужен другой SMTP-порт хоста, поменяйте его слева в mapping и в `SMTP_ADDRESS`. `SMTP_ALLOW_INSECURE=true` здесь разрешает локальный Mailpit без TLS. Для внешнего SMTP используйте его адрес, отправителя, реквизиты и поддержку STARTTLS; параметры приведены ниже.

Go-процесс останавливается `Ctrl+C`. Вернуться к контейнерному API можно командой `make up`. Если временное SMTP-переопределение больше не нужно, пересоздайте Mailpit обычным `docker compose up -d mailpit` и удалите созданный временный YAML. Данные приложения находятся в PostgreSQL и этим действием не удаляются.

## 7 Настройки и их источники

`internal/config` читает `config/config.yaml`, применяет ENV и проверяет получившиеся типизированные настройки. `CONFIG_PATH` выбирает другой YAML. Относительный путь `migrations.directory` считается от директории YAML. Секреты БД и SMTP передаются через окружение. Перезапустите API после изменения настроек.

`.env` используется Make и для подстановок Compose. Наличие переменной в `.env` само по себе не передаёт её контейнеру: учитываются только настройки, перечисленные в `compose.yaml`. В частности, Compose явно задаёт `SMTP_ADDRESS=mailpit:1025`, отправителя, CORS и HTTP-адрес. Для изменения этих значений в контейнерном варианте измените соответствующий `api.environment` в Compose или создайте собственное переопределение; для host Go используйте переменные ниже.

| Параметр | Назначение и значение по умолчанию |
| --- | --- |
| `CONFIG_PATH` | YAML приложения; по умолчанию `config/config.yaml` |
| `DB_USER`, `DB_PASSWORD` | Обязательные реквизиты PostgreSQL; из окружения |
| `DB_HOST`, `DB_PORT` | Переопределение адреса БД; YAML: `localhost:5433`, Compose: `postgres:5432` |
| `database.name`, `database.ssl_mode` | Имя БД `tourneyhub` и SSL `disable`; задаются YAML |
| `HTTP_ADDRESS` | Адрес Go HTTP-сервера; по умолчанию `:8080` |
| `CORS_ALLOWED_ORIGINS` | Разрешённые origins через запятую; YAML: порты frontend 5173 и Swagger 8081 |
| `PUBLIC_WEB_URL` | Origin frontend для ссылок из писем и приглашений; `http://localhost:5173` |
| `SMTP_ADDRESS`, `SMTP_FROM` | SMTP `host:port` и отправитель; host YAML: пустой адрес и `noreply@tourneyhub.local` |
| `SMTP_USER`, `SMTP_PASSWORD` | Реквизиты SMTP, если сервер требует аутентификацию |
| `SMTP_ALLOW_INSECURE` | Разрешение SMTP без TLS; по умолчанию `false`, локальный Mailpit в Compose: `true` |
| `SMTP_DIAL_TIMEOUT`, `SMTP_TIMEOUT` | Таймаут подключения и операции SMTP; `10s`, `15s` |
| `WORKER_INTERVAL`, `EVENT_RETENTION` | Период фоновых задач `1m` и хранение SSE-истории `168h` |
| `HTTP_READ_HEADER_TIMEOUT`, `HTTP_READ_TIMEOUT` | HTTP-таймаут заголовков `5s` и чтения `15s` |
| `HTTP_IDLE_TIMEOUT`, `HTTP_SHUTDOWN_TIMEOUT` | Idle `60s` и корректная остановка `10s` |
| `HTTP_BODY_LIMIT` | Ограничение тела запроса; `6M` |
| `SSE_POLL_INTERVAL`, `SSE_HEARTBEAT_INTERVAL` | Опрос событий `1s` и heartbeat `15s` |
| `POSTGRES_PORT`, `API_PORT`, `SWAGGER_PORT`, `MAILPIT_PORT` | Публикуемые порты Compose; `5433`, `8080`, `8081`, `8025` |
| `ADMIN_EMAIL`, `ADMIN_NICKNAME`, `ADMIN_PASSWORD` | Параметры отдельной команды bootstrap |

Для другого frontend-origin учитывайте оба назначения: **CORS** разрешает браузерные запросы, а **PUBLIC_WEB_URL** формирует ссылки. Одно не заменяет другое. Все даты API передаются в RFC3339; часовой пояс турнира — IANA, например `Europe/Moscow`.

## 8 Миграции, проверки и метрики

Для миграций контейнерного стенда:

```sh
make migrate-status
make migrate-up
```

`make up` уже выполняет `up`; повтор применяет только новые миграции. `make migrate-down` откатывает последнюю миграцию и изменяет схему, поэтому не используйте его для обычного перезапуска или исправления доступа к БД.

Для разработки с Go на хосте:

```sh
make generate-check
make check
make build
make docs
make metrics
```

`make check` уже включает проверку воспроизводимости обоих generated-файлов, обычные тесты и `go vet`. `make generate` заново создаёт OpenAPI-код и DTO mapping; сгенерированные файлы не редактируются вручную. `make build` создаёт API, migrate и bootstrap в `bin/`. `make docs` выгружает GoDoc модели в `build/docs/example.txt`; дополнительный HTML/RTF-экспорт `make docs-export` требует Pandoc.

Результаты метрик сохраняются в `build/metrics/current.json`, `gocyclo.txt`, `gocognit.txt`. Первоначальные и повторные данные ЛР3, правила отбора и архивы исходников находятся в [разделе ЛР3](../docs/lab3/README.md). Не перезаписывайте первоначальный замер результатом текущего кода.

### Интеграционные тесты в отдельной PostgreSQL

Интеграционные тесты требуют PostgreSQL и `TEST_DATABASE_URL`. Тест миграций выполняет `up/down`; **адрес рабочей БД приложения использовать нельзя**. Создайте отдельный временный контейнер без общего volume. Для `-race` также нужна поддерживаемая платформа и C toolchain, например Xcode Command Line Tools на macOS или GCC на Linux.

```sh
docker run --rm -d --name tourneyhub-integration-postgres \
  -e POSTGRES_DB=tourneyhub_test \
  -e POSTGRES_USER=tourneyhub_test \
  -e POSTGRES_PASSWORD=integration_only_2026 \
  -p 127.0.0.1:55433:5432 postgres:18.6-alpine

until docker exec tourneyhub-integration-postgres \
  pg_isready -U tourneyhub_test -d tourneyhub_test; do
  sleep 1
done

TEST_DATABASE_URL='postgres://tourneyhub_test:integration_only_2026@localhost:55433/tourneyhub_test?sslmode=disable' \
  make integration-test
```

Если порт 55433 занят, выберите другой в `docker run` и DSN. Эти реквизиты относятся только к созданному тестовому контейнеру. Если вместо обычного прогона нужен общий профиль покрытия, выполните до остановки контейнера:

```sh
mkdir -p build
TEST_DATABASE_URL='postgres://tourneyhub_test:integration_only_2026@localhost:55433/tourneyhub_test?sslmode=disable' \
  go test -tags=integration -race -count=1 -coverpkg=./internal/... \
  -coverprofile=build/coverage.out ./...
```

Профиль учитывает прохождение внутренних слоёв из разных тестовых пакетов. После проверок остановите именно временный тестовый контейнер:

```sh
docker stop tourneyhub-integration-postgres
```

Из-за `--rm` контейнер удалится автоматически. Основной Compose и его volume эта команда не затрагивает.

## 9 Перезапуск и остановка

```sh
docker compose restart api
make logs
make down
```

`restart api` перезапускает тот же контейнер; он не пересобирает исходники и не применяет новые настройки окружения. После изменения кода или `.env` используйте `make up`: образ пересоберётся, миграции применятся, контейнеры при необходимости пересоздадутся.

`make down` останавливает и удаляет контейнеры/сеть проекта, но сохраняет named volume PostgreSQL. Следующий `make up` использует прежние данные. Не добавляйте `-v` и не удаляйте volume, если нужны созданные аккаунты, турниры и файлы.

## 10 Типичные проблемы

| Симптом | Что проверить |
| --- | --- |
| `docker compose` не выполняется | Docker запущен, установлен Compose v2; терминал находится в `backend/` |
| Порт занят | Измените соответствующий `*_PORT` в `.env`; host API отдельно использует `HTTP_ADDRESS` |
| `DB_USER и DB_PASSWORD обязательны` | `.env` существует; использован Make/Compose либо переменные экспортированы для прямого Go-процесса |
| `password authentication failed` | Реквизиты совпадают с ролью в уже существующем volume; редактирование `.env` не переинициализирует PostgreSQL |
| Миграции завершились ошибкой | `docker compose logs migrate`, доступность PostgreSQL, правильные YAML и каталог миграций; не запускайте API в обход неуспешных миграций |
| Письмо не появилось сразу | Подождите один период worker, откройте правильный Mailpit, проверьте `docker compose logs --tail=100 api`; для host Go настройте опубликованный SMTP-порт |
| Ссылка подтверждения не открывается | Frontend на `PUBLIC_WEB_URL` ещё не запущен; возьмите `token` из письма и вызовите `/auth/verify-email` вручную |
| Login отклонён | E-mail подтверждён, пароль верен, аккаунт не заблокирован; после серии попыток дождитесь окончания минутного лимита |
| Bootstrap отказался создать второго администратора | Первый уже существует; войдите им и назначьте роли зарегистрированному пользователю через API |
| Каталог дисциплин/турниров пуст | На новом стенде это ожидаемо: администратор создаёт дисциплины, организатор — турниры |
| Browser CORS error | Совпадает полный origin frontend с разрешённым CORS; учитывайте `api.environment` в Compose |
| Swagger не загрузил спецификацию | API готов, `API_PORT` верен, открыт `/openapi.yaml`; после изменения портов пересоздан Swagger |
| Обновление вернуло 409 | Перечитайте объект и используйте актуальную `version`; не повторяйте старое изменение бесконечно |

Описание внутреннего устройства backend находится в [архитектурном документе](../docs/backend-architecture.md). Краткая передача frontend-разработчику — в [frontend-handoff.md](../docs/lab3/frontend-handoff.md).
