# Сервисная архитектура системы управления пользователями

## Сервисы системы

### 1. API-Service
**Назначение**: Обрабатывает входящие запросы, выполняет маршрутизацию и базовую авторизацию через SessionManagement.

**Основные функции**:
- **Middleware**:
  - Logging (traceID, таймауты)
  - Rate-Limiter
  - Authorization через gRPC-запросы в SessionManagement
- **Проксирование**:
  - `/auth/register`, `/auth/login`, `users/logout`, `users/del` `/me/update` → UserManagement
  - `/me`, `/users/{id}` → UserManagement + Photo_Service
  - `/me/photos`, `/me/photos{photo_id}`, `/users/{id}/photos/{photo_id}` → Photo_Service
- gRPC-запросы в Photo-Service
- Документация Swagger
- Отправка логов в Logs-Service в формате JSON
- Отправка метрик в Prometheus

**Технологии**: 
`Gin` | `gRPC-Client` | `Swagger` | `Kafka-Producer` | `Prometheus`

---

### 2. UserManagement
**Назначение**: Управление профилями пользователей.

**Эндпоинты**:
- `POST /auth/register`
- `POST /auth/login`
- `DELETE /users/logout` 
- `DELETE /users/delete`
- `PATCH /me/update`
- `GET /me`
- `GET /users/{id}`

**Бизнес-логика**:
- RegistrateAndLogin
- AuthenticateAndLogin
- Logout
- DeleteAccount
- UpdateAccount
- GetProfileById

**Особенности**:
- Использование миграций баз данных
- Отправка метрик в Prometheus
- Отправка событий в Photo-Service
- Кэширование профилей пользователей
- Отправка логов в Logs-Service в формате JSON

**Технологии**: 
`net/http` | `PostgreSQL` | `gRPC-Client` | `bcrypt` | `Kafka-Producer` | `Prometheus` | `Redis`

---

### 3. SessionManagement
**Назначение**: Управление сессиями пользователей.

**gRPC методы**:
- CreateSession
- ValidateSession
- DeleteSession

**Особенности**:
- Redis с TTL
- Отправка логов в Logs-Service в формате JSON

**Технологии**: 
`gRPC-Server` | `Redis` | `Kafka-Producer`

---

### 4. Logs-Service
**Назначение**: Централизованный сбор логов со всех сервисов.

**Технологии**: 
`Kafka-Consumer` | `Zap-Logger` | `os`

---

### 5. Photo-Service
**Назначение**: Управление фотографиями пользователей.

**gRPC методы**:
- LoadPhoto
- DeletePhoto
- GetPhoto
- GetPhotos

**RabbitConsumer события**:
- user.registration
- user.delete

**Особенности**:
- Использование миграций баз данных
- Хранение ID пользователей и метаданных фотографий в PostgreSQL (каскадное удаление)
- Кэширование метаданных фотографий пользователей с TTL
- Интеграция с хранилищем Mega.nz
- Получение событий от UserManagement
- Отправка логов в Logs-Service

**Технологии**: 
`gRPC-Server` | `PostgreSQL` | `Rabbit-Consumer` | `Mega.nz` | `Redis`

---

## Общая архитектура

### Конфигурация
- Личный `config.yml` (Viper) для каждого сервиса

### Логирование
- Стандартная библиотека/Zap
- Централизованный сбор через Logs-Service

### Ключевые характеристики:
- Модульная структура (разделение на пакеты)
- Retry-логика для внешних вызовов
- Обработка таймаутов
- Graceful shutdown во всех сервисах
- Разделение ошибок на клиентские/серверные