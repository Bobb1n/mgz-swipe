# swipe-mgz

Микросервис свайпов, матчей и геолокации. Часть dating-платформы.

## Возможности

- **Свайпы** (like/dislike): `POST /v1/swipes`
- **Взаимный лайк → match**: автоматическое создание записи в `matches`
- **Список матчей**: `GET /v1/matches?limit&offset`
- **Геолокация**: `PUT /v1/location` (широта/долгота) + `GET /v1/candidates` (кандидаты в радиусе N км)
- **Kafka events**: `swipe.events`, `match.events`

## Архитектура

- HTTP (Echo) — фасад для API Gateway
- PostgreSQL — таблицы `swipes`, `matches`
- Redis Geo (DB 1) — координаты пользователей, поиск в радиусе
- Kafka (`segmentio/kafka-go`) — асинхронная публикация событий
- Clean Architecture: usecase зависит от интерфейсов (`SwipeRepository`, `MatchRepository`, `LocationRepository`, `EventPublisher`)

```
cmd/             — entry point
internal/
  config/        — конфиг из env
  domain/        — модели (Swipe, Match, Candidate)
  repository/    — Postgres + Redis
  events/        — Kafka publisher
  usecase/       — бизнес-логика + интерфейсы
  transport/http — HTTP handlers
migrations/      — goose/migrate SQL
swagger.yaml     — OpenAPI 3.0
```

## Локальный запуск

Сервис поднимается вместе с остальными через `docker-compose.yml` из `auth_service`:

```bash
cd ../auth_service
docker compose up -d
```

Переменные окружения (см. `.env.example`):

| Var | Default | Описание |
|---|---|---|
| `SERVER_PORT` | `8084` | HTTP порт |
| `DATABASE_URL` | `postgres://...swipe_service` | Postgres DSN |
| `REDIS_ADDR` | `redis:6379` | Redis для geo |
| `REDIS_DB` | `1` | Отдельная DB для geo |
| `KAFKA_BROKERS` | `kafka:9092` | Брокеры |
| `GEO_RADIUS_KM` | `50` | Радиус по умолчанию |

## HTTP API

Полная спецификация — `swagger.yaml`.

| Метод | Путь | Описание |
|---|---|---|
| PUT | `/v1/location` | Обновить координаты |
| GET | `/v1/candidates` | Кандидаты в радиусе |
| POST | `/v1/swipes` | Like/dislike |
| GET | `/v1/matches` | Список матчей |

Все защищённые эндпоинты получают идентификатор пользователя из заголовка `X-User-Id`, который проставляет API Gateway после валидации JWT.

## Kafka топики

**`swipe.events`**
```json
{"swiper_id": "...", "swipee_id": "...", "direction": "like", "created_at": "..."}
```

**`match.events`**
```json
{"match_id": 3, "user1_id": "...", "user2_id": "...", "created_at": "..."}
```
