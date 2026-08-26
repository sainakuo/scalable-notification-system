#  Scalable Notification System

A scalable backend system for asynchronous task processing built with Go.

The project demonstrates a production-like backend architecture with REST API, PostgreSQL, Redis, background workers, gRPC communication, retry handling, observability, health checks, graceful shutdown, and automated testing.

---

#  Features

- REST API with Gin
- PostgreSQL persistent storage
- Redis-based task queue
- Background task processing
- Worker Pool using goroutines and channels
- Retry strategy for failed tasks
- Idempotent task processing
- gRPC communication between services
- Rollback when queue publishing fails
- Structured logging with `log/slog`
- Prometheus metrics
- PostgreSQL and Redis health checks
- Graceful Shutdown
- Dependency inversion through interfaces
- Application Composition Root
- Database indexes and constraints
- Unit tests
- Integration tests with PostgreSQL and Redis
- Docker Compose deployment
- Swagger/OpenAPI documentation
- Configuration via environment variables
- Database migrations

---

#  Architecture

```text
                           +----------------+
                           |     Client     |
                           +-------+--------+
                                   |
                              HTTP REST API
                                   |
                                   v
                        +--------------------+
                        |      Gin API       |
                        +---------+----------+
                                  |
                             TaskService
                                  |
                    +-------------+-------------+
                    |                           |
                    v                           v
             +-------------+             +-------------+
             | PostgreSQL  |             | Redis Queue |
             +-------------+             +------+------+
                                               |
                                               v
                                    +--------------------+
                                    |    Worker Pool     |
                                    |   5 goroutines     |
                                    +---------+----------+
                                              |
                                       Retry Strategy
                                              |
                                              v
                                     +----------------+
                                     |  gRPC Sender   |
                                     +----------------+
```

The API stores task data in PostgreSQL and publishes the task ID to Redis.

Workers consume task IDs from Redis, retrieve task data from PostgreSQL, process notifications through the gRPC sender service, and update task status.

Tasks move through the following lifecycle:

```text
pending → processing → done
                     ↘ retry → failed
```

---

# Reliability

The system includes several mechanisms designed to make asynchronous task processing more reliable.

### Retry Strategy

When notification processing fails, the worker checks whether the task can be retried.

```text
processing
    ↓
send error
    ↓
ShouldRetry?
   /       \
 yes       no
  ↓         ↓
retry     failed
  ↓
Redis Queue
```

The retry counter is stored in PostgreSQL.

### Idempotent Processing

Before processing a task, the worker checks its current status.

Tasks already marked as:

```text
done
failed
```

are skipped.

This prevents the same completed task from being processed again if its ID appears in the queue more than once.

### Queue Publishing Rollback

Task creation involves two systems:

```text
PostgreSQL → Redis
```

If the task is successfully stored in PostgreSQL but publishing its ID to Redis fails, the service attempts to remove the created database record.

This prevents tasks from remaining permanently in `pending` without ever reaching a worker.

---

# Observability

The project includes basic observability mechanisms for monitoring application state and diagnosing failures.

### Structured Logging

Application and worker events use Go's `log/slog` with structured fields such as:

```text
task_id
worker_id
user_id
task_type
error
```

### Prometheus Metrics

The worker exposes Prometheus-compatible metrics on:

```text
http://localhost:9090/metrics
```

Metrics include task processing, failures, retries, processing duration, and queue size.

### Health Checks

```http
GET /health
```

checks application dependencies including:

- PostgreSQL
- Redis

If a required dependency is unavailable, the endpoint returns an unhealthy response.

---

#  Tech Stack

- Go
- Gin
- PostgreSQL
- Redis
- gRPC / Protocol Buffers
- Prometheus
- `log/slog`
- Docker
- Docker Compose
- Swagger / OpenAPI
- Goroutines
- Channels
- pgx
- go-redis

Concepts and practices:

- REST API
- Worker Pool
- Repository Pattern
- Dependency Injection
- Dependency Inversion
- Composition Root
- Structured Logging
- Metrics
- Health Checks
- Observability
- Graceful Shutdown
- Unit Testing
- Integration Testing

---

#  Project Structure

```text
cmd/
├── api/
├── worker/
└── grpc-sender/

internal/
├── app/
├── config/
├── handler/
├── health/
├── logger/
├── metrics/
├── model/
├── queue/
├── repository/
├── sender/
├── service/
└── worker/

proto/
├── notification.proto
└── notificationpb/

migrations/
docs/

docker-compose.yml
```

`internal/app` acts as the application Composition Root and is responsible for wiring application dependencies.

---

#  Getting Started

## Clone repository

```bash
git clone https://github.com/sainakuo/scalable-notification-system.git
cd scalable-notification-system
```

## Run with Docker

```bash
docker compose up --build
```

The following services will start:

- API
- Worker
- gRPC Sender
- PostgreSQL
- Redis
- Database migrations

---

#  API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Check application dependencies |
| POST | `/tasks` | Create a task |
| GET | `/tasks` | Get all tasks |
| GET | `/tasks/{id}` | Get task by ID |

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

---

#  Create Task

### curl

```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "user_id": 1,
    "type": "email",
    "payload": "Welcome!"
}'
```

### PowerShell

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/tasks" `
-Method POST `
-ContentType "application/json" `
-Body '{"user_id":1,"type":"email","payload":"Welcome!"}'
```

---

# Database

Tasks are stored in PostgreSQL.

The `tasks` table contains:

```text
id
user_id
type
payload
status
retry_count
created_at
```

Indexes and database constraints are used to improve query performance and protect data integrity.

Connect to PostgreSQL:

```bash
docker exec -it sns_postgres psql -U postgres -d sns_db
```

Example:

```sql
SELECT id,
       type,
       status,
       retry_count
FROM tasks
ORDER BY id DESC;
```

---

# Testing

The project contains both unit and integration tests.

## Unit Tests

Unit tests verify application logic without requiring PostgreSQL, Redis, or the gRPC service.

Covered components include:

- TaskService
- Worker processing
- Retry behavior
- Failed task handling
- Idempotent task processing
- Queue publishing rollback

Run:

```bash
go test ./...
```

## Integration Tests

Integration tests verify application components against real PostgreSQL and Redis instances.

Covered scenarios include:

```text
TaskRepository → PostgreSQL
RedisQueue → Redis
TaskService → PostgreSQL + Redis
```

Enable integration tests:

### PowerShell

```powershell
$env:INTEGRATION_TEST="1"
go test ./...
```

Integration tests require PostgreSQL and Redis to be running.

> The worker should be stopped while Redis queue integration tests are running because both the test and worker consume the same queue.

After testing:

```powershell
Remove-Item Env:INTEGRATION_TEST
```

---

# What This Project Demonstrates

This project demonstrates practical backend development experience with:

- Designing REST APIs in Go
- Layered backend architecture
- PostgreSQL persistence
- Redis queues
- Asynchronous processing
- Worker pools and concurrency
- Retry strategies
- Idempotent processing
- gRPC service communication
- Error propagation and domain errors
- Dependency inversion
- Graceful shutdown
- Structured application logging
- Prometheus metrics
- Dependency health checks
- Database indexes and constraints
- Unit and integration testing
- Dockerized multi-service applications

---

# Roadmap

Completed:

- REST API and PostgreSQL persistence
- Redis task queue
- Worker Pool
- gRPC notification sender
- Retry strategy
- Graceful Shutdown
- Idempotent processing
- Structured logging
- Health checks
- Prometheus metrics
- Database indexes and constraints
- Unit tests
- Integration tests

Next:

- Load testing with `wrk`
- Performance analysis and documentation