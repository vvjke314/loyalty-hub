# LoyaltyHub

Сервис лояльности для управления заказами, балансом пользователей и системой начисления бонусов.

## Описание проекта

LoyaltyHub - это REST API сервис, построенный на архитектуре Clean Architecture с разделением на слои:
- **Handlers** - HTTP обработчики (Gin)
- **Services** - бизнес-логика
- **Repository** - работа с базой данных
- **Models** - доменные модели

### Основные возможности:
- Регистрация и аутентификация пользователей
- Загрузка и отслеживание заказов
- Управление балансом и выводами средств
- Интеграция с внешним сервисом начисления бонусов
- Система JWT токенов с refresh механизмом

## Технологии

### Backend:
- **Go 1.24.2** - основной язык
- **Gin** - HTTP фреймворк
- **PostgreSQL** - основная база данных
- **pgx** - драйвер PostgreSQL
- **JWT** - аутентификация
- **bcrypt** - хеширование паролей

### Observability:
- **OpenTelemetry** - трейсинг
- **Jaeger** - визуализация трейсов
- **Zap** - структурированное логирование
- **Prometheus** - метрики

### Документация:
- **Swagger/OpenAPI** - API документация
- **Swaggo** - генерация документации

### Дополнительные:
- **Docker & Docker Compose** - контейнеризация
- **Redis** - кеширование
- **Circuit Breaker** - устойчивость к сбоям
- **Rate Limiting** - ограничение запросов

## Структура проекта

```
loyalityhub/
├── cmd/                    # Точка входа приложения
├── internal/               # Внутренний код
│   ├── app/               # Инициализация приложения
│   ├── handlers/          # HTTP обработчики
│   ├── services/          # Бизнес-логика
│   ├── repository/        # Работа с БД
│   ├── model/             # Доменные модели
│   ├── dto/               # Data Transfer Objects
│   ├── middleware/        # HTTP middleware
│   ├── router/            # Настройка маршрутов
│   ├── utils/             # Утилиты
│   ├── client/            # HTTP клиенты
│   └── tracing/           # Настройка трейсинга
├── migrations/            # Миграции БД
├── docs/                  # Swagger документация
├── docker-compose.yaml    # Docker Compose конфигурация
└── Makefile              # Команды для сборки и запуска
```

## Установка и запуск

### Предварительные требования:
- Go 1.24.2+
- Docker & Docker Compose
- PostgreSQL (если запуск без Docker)

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd loyalityhub
```

### 2. Настройка переменных окружения
```bash
cp .env_example .env
```

Отредактируйте `.env` файл под ваши настройки:
```env
# Основные настройки приложения
APP_HOST=localhost

# База данных
ORDERS_DB_PASS=supersecretpass
ORDERS_DB_USER=superuser
ORDERS_DB_PORT=5432
ORDERS_DB_NAME=orders
ORDERS_DB_DSN=postgres://superuser:supersecretpass@orders_db:5432/orders?sslmode=disable
ORDERS_DB_DSN_TEST=postgres://superuser:supersecretpass@localhost:5432/orders?sslmode=disable

# Логирование
LOG_FILE=./app.log

# JWT
JWT_SECRET=supersecretjwt

# Observability
JAEGER_LISTEN_HOST_TEST=localhost
JAEGER_LISTEN_HOST=jaeger
JAEGER_LISTEN_PORT=4318

# Внешний сервис начисления
ACCRUAL_SERVICE=localhost:8090
```

### 3. Запуск с помощью Makefile

#### Полный запуск (с внешним сервисом начисления):
```bash
make run-with-accrual
```

#### Запуск без внешнего сервиса:
```bash
make run-without-acrrual
```

#### Сборка Docker образов:
```bash
make build
```

#### Очистка контейнеров:
```bash
make container-rm
```

#### Очистка volumes:
```bash
make volume-rm
```

### 4. Запуск вручную

#### С Docker Compose:
```bash
docker-compose up -d
```

#### Локально:
```bash
go mod tidy
go run cmd/main.go
```

## API Endpoints

### Аутентификация

#### Регистрация пользователя
```http
POST /api/v1/register
Content-Type: application/json

{
  "login": "user@example.com",
  "password": "password123"
}
```

#### Вход в систему
```http
POST /api/v1/auth
Content-Type: application/json

{
  "login": "user@example.com",
  "password": "password123"
}
```

#### Обновление access токена
```http
GET /api/v1/refresh
```

### Заказы (требуют аутентификации)

#### Загрузка заказа
```http
POST /api/v1/user/orders
Authorization: Bearer <access_token>
Content-Type: text/plain

1234567890
```

#### Получение списка заказов
```http
GET /api/v1/user/orders
Authorization: Bearer <access_token>
```

### Баланс (требуют аутентификации)

#### Получение баланса
```http
GET /api/v1/user/balance
Authorization: Bearer <access_token>
```

#### Вывод средств
```http
POST /api/v1/user/balance/withdraw
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "order": "1234567890",
  "sum": 100.50
}
```

#### История выводов
```http
GET /api/v1/user/withdrawals
Authorization: Bearer <access_token>
```

## Документация API

Swagger UI доступен по адресу: `http://localhost:8080/swagger/index.html`

Для обновления документации:
```bash
swag init -g cmd/main.go
```

## Логирование

Логи записываются в файл `app.log` в корне проекта. Настройки логирования:
- **Уровень**: Debug (консоль), Info (файл)
- **Формат**: JSON (файл), Console (консоль)
- **Ротация**: Ручная (требует перезапуска)

## Мониторинг

### Метрики Prometheus
```http
GET /metrics
```

### Трейсинг Jaeger
- **URL**: `http://localhost:16686`
- **Экспорт**: OTLP HTTP на порту 4318

## Разработка

### Структура кода
Проект следует принципам Clean Architecture:
- **Handlers** - обработка HTTP запросов
- **Services** - бизнес-логика
- **Repository** - доступ к данным
- **Models** - доменные сущности

### Добавление новых endpoints
1. Создать handler в `internal/handlers/`
2. Добавить бизнес-логику в `internal/services/`
3. Создать репозиторий в `internal/repository/`
4. Зарегистрировать маршрут в `internal/router/`
5. Добавить swagger комментарии

### Тестирование
```bash
go test ./...
```

## Лицензия

MIT License
