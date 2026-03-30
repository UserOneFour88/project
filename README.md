# Concurrent Pipeline (Go)

Консольное приложение, которое:
- получает данные из API `https://jsonplaceholder.typicode.com/posts`;
- обрабатывает их конкурентным pipeline;
- использует `fan-out/fan-in`;
- выводит форматированный результат в `stdout`.

## Требования

- Go 1.22+ (или совместимая версия)

Проверить установку:

```bash
go version
```

## Запуск

Из корня проекта:

```bash
go run ./cmd/pipeline -limit 5 -user 1 -workers 4
```

## Флаги

- `-user` — фильтр по `userId` (по умолчанию `0`, то есть все пользователи)
- `-limit` — максимум постов в обработке (`0` = без лимита)
- `-workers` — количество конкурентных воркеров для fan-out
- `-timeout` — timeout HTTP-запроса (например, `5s`, `10s`)

Пример:

```bash
go run ./cmd/pipeline -user 2 -limit 7 -workers 3 -timeout 8s
```

## Архитектура и границы ответственности

- `internal/api` — HTTP-клиент и получение данных из внешнего API
- `internal/model` — доменные структуры (`Post`, `PostSummary`)
- `internal/pipeline` — этапы pipeline + fan-out/fan-in
- `internal/format` — форматирование выходных данных
- `cmd/pipeline` — CLI, флаги, сборка этапов, обработка ошибок и вывод

## Pipeline

1. **Source**: `[]Post -> chan Post`
2. **FilterAndMap**: фильтрует по `userId`, отрезает по `limit`, подготавливает `PostSummary`
3. **Fan-out**: `N` воркеров параллельно считают слова в `body`
4. **Fan-in**: слияние результатов всех воркеров в один канал
5. **Format + stdout**: сортировка и печать строк

## Формат вывода

Каждая строка:

```text
post_id=<id> user_id=<userId> words=<body_words_count> title="<title>"
```

## Обработка ошибок

Программа завершится с кодом `1` и сообщением в `stderr`, если:
- переданы невалидные значения флагов;
- API недоступен или вернул неожиданный статус;
- JSON не удалось декодировать;
- после фильтрации не осталось данных.

---

## Минимальный чат-сервер (TCP)

В проекте также есть простой конкурентный чат-сервер:

- файл: `cmd/chatserver/main.go`
- модель: `hub` + каналы `register/unregister/broadcast`
- рассылка сообщений всем подключенным клиентам (fan-out)

### Запуск чат-сервера

```bash
go run ./cmd/chatserver -addr :9000
```

### Подключение клиентов

Можно открыть 2+ терминала и подключиться:

```bash
telnet 127.0.0.1 9000
```

или (если установлен `nc`):

```bash
nc 127.0.0.1 9000
```

### Использование

1. Введите имя при подключении.
2. Пишите сообщения и нажимайте Enter.
3. Для выхода введите `/quit`.

### Команды чата

- `/join <room>` — перейти в комнату
- `/msg <user> <text>` — отправить личное сообщение пользователю
- `/who` — список пользователей онлайн
- `/quit` — выйти из чата

Примеры:

```text
/join backend
/msg alice привет, глянь pipeline
/who
```

---

## WebSocket чат (HTTP + gorilla/websocket)

Требования задания:
- простой HTTP-сервер с endpoint’ом;
- WebSocket через `github.com/gorilla/websocket`;
- комнаты (room);
- чтение/запись сообщений в комнату;
- простой клиент;
- идентификация пользователя через “auth-токен” (не безопасность, а идентификатор).

Файлы:
- сервер: `cmd/wsserver/main.go`
- клиент: `cmd/wsclient/main.go`

### Запуск сервера

```bash
go run ./cmd/wsserver -addr :8081
```

HTTP endpoint:
- `GET /health` → `ok`

WebSocket endpoint:
- `GET /ws`

### Протокол (очень простой JSON)

1) Клиент подключается к `ws://localhost:8081/ws` и отправляет `join`:

```json
{"type":"join","room":"lobby","name":"alice","token":""}
```

2) Сервер отвечает “token”:

```json
{"type":"token","room":"lobby","name":"alice","token":"<issued-token>"}
```

3) Дальше клиент шлёт сообщения:

```json
{"type":"msg","text":"hello"}
```

4) Сервер рассылает всем в комнате:

```json
{"type":"msg","room":"lobby","name":"alice","text":"hello"}
```

### Запуск клиента

Подключиться и получить токен:

```bash
go run ./cmd/wsclient -url ws://localhost:8081/ws -room lobby -name alice
```

Подключиться уже с токеном:

```bash
go run ./cmd/wsclient -url ws://localhost:8081/ws -room lobby -token <token>
```

---

## REST API + Auth + Postgres (миграции) + DI

Это следующая часть проекта: REST, JWT (access/refresh), Postgres и миграции. DI реализован без контейнера — зависимости собираются в `cmd/api/main.go` через конструкторы.

Файлы:
- REST сервер: `cmd/api/main.go`
- Роуты/handlers: `internal/httpapi/httpapi.go`
- JWT: `internal/auth/jwt.go`
- Postgres repo: `internal/store/*`
- Конфиг из env: `internal/config/config.go`
- Миграции: `migrations/*.sql`
- Инфраструктура: `.env.example`, `Dockerfile.api`, `docker-compose.yml`

### Запуск (локально, без Docker)

1) Подними Postgres (любой способ) и создай БД `app` с юзером `app/app`.

2) Примени миграции:

- **То, что я НЕ могу сделать за тебя автоматически**: применить миграции на твоей машине без инструментов.
- **Как сделать самому (варианты)**:
  - поставить `migrate` (golang-migrate) и выполнить:

```bash
migrate -path ./migrations -database "postgres://app:app@localhost:5432/app?sslmode=disable" up
```

  - или открыть `migrations/0001_init.up.sql` и выполнить SQL в твоём клиенте Postgres.

3) Запусти API (env можно взять из `.env.example`):

```bash
set HTTP_ADDR=:8082
set DATABASE_URL=postgres://app:app@localhost:5432/app?sslmode=disable
set JWT_ACCESS_SECRET=change-me-access
set JWT_REFRESH_SECRET=change-me-refresh
go run ./cmd/api
```

Health:
- `GET http://localhost:8082/health`

### Запуск (Docker)

1) Сборка и старт контейнеров:

```bash
docker compose up --build
```

2) Применить миграции (вариант через `migrate` на хосте):

```bash
migrate -path ./migrations -database "postgres://app:app@localhost:5432/app?sslmode=disable" up
```

### REST endpoints (минимум)

Auth:
- `POST /auth/register` `{ "username": "...", "password": "..." }`
- `POST /auth/login` `{ "username": "...", "password": "..." }` → `{ access_token, refresh_token }`
- `POST /auth/refresh` `{ "refresh_token": "..." }` → новые токены (rotation)
- `POST /auth/logout` `{ "refresh_token": "..." }` → revoke refresh

Protected (Bearer access token):
- `GET /me`
- `GET /rooms`
- `POST /rooms` `{ "name": "..." }`
- `GET /rooms/{roomID}/messages?limit=50`
- `POST /rooms/{roomID}/messages` `{ "text": "..." }`
