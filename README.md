# Scalable Notification System

A scalable backend system for asynchronous task processing built with Go.

The project demonstrates a production-like backend architecture with REST API, PostgreSQL, Redis, background workers, gRPC communication, retry handling, idempotency protection, observability, TLS support, health checks, graceful shutdown, and automated testing.

---

# Features

- REST API with Gin
- PostgreSQL persistent storage
- Redis-based task queue
- Background task processing
- Worker Pool using goroutines and channels
- Retry strategy with persistent retry counters
- Idempotency protection for notification delivery
- gRPC communication between services
- Optional TLS for gRPC communication
- Compensating rollback when queue publishing fails
- Structured logging with `log/slog`
- Prometheus metrics
- OpenTelemetry distributed tracing
- PostgreSQL and Redis health checks
- Graceful Shutdown
- Dependency inversion through interfaces
- Application Composition Root
- Database indexes and constraints
- Unit tests
- Integration tests with PostgreSQL and Redis
- Load testing with `wrk`
- Docker Compose deployment
- Swagger/OpenAPI documentation
- Configuration via environment variables
- Database migrations

---

# Architecture

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
                                    |    5 goroutines    |
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
              |
              v
            retry
              |
              v
            failed
```

The application is separated into transport, service, persistence, queue, worker, sender, metrics, tracing, and health-check layers.

`internal/app` acts as the Composition Root and wires concrete implementations together.

---

# Reliability

The system includes several mechanisms designed to make asynchronous task processing more reliable.

## Retry Strategy

When notification processing fails, the worker checks whether the task can be retried.

```text
processing
    |
    v
send error
    |
    v
ShouldRetry?
   /       \
 yes        no
  |          |
  v          v
retry      failed
  |
  v
Redis Queue
```

The retry counter is stored in PostgreSQL, so retry state survives worker restarts.

After the configured retry limit is reached, the task is moved to the terminal `failed` state.

---

## Idempotency Protection

Workers skip tasks that are already in a terminal state:

```text
done
failed
```

The gRPC notification service also stores processed task IDs in Redis:

```text
notification:processed:<task_id>
```

Processed IDs have a TTL and are checked before notification delivery.

This protects against common duplicate-delivery scenarios, including a task being requeued after the notification has already been successfully sent.

The project uses **at-least-once processing with idempotency protection** rather than claiming strict exactly-once delivery.

A small concurrency window still exists because checking whether an ID was processed and marking it as processed are separate Redis operations. A production system requiring stronger guarantees could use an atomic reservation strategy or a more advanced delivery state machine.

---

## Queue Publishing Rollback

Task creation involves two independent systems:

```text
PostgreSQL → Redis
```

The service first stores the task in PostgreSQL and then publishes its ID to Redis.

If Redis publishing fails after the database insert succeeds, the service performs a compensating rollback by deleting the newly created database record.

```text
INSERT PostgreSQL
       |
       v
  LPUSH Redis
       |
     error
       |
       v
DELETE PostgreSQL
```

This prevents tasks from remaining permanently in `pending` without ever reaching a worker.

This is intentionally a simpler approach than the Transactional Outbox pattern. In a larger production system, an Outbox would provide stronger consistency guarantees between database state and event publishing.

---

# gRPC Security

Communication between the Worker and gRPC Sender supports both plaintext and TLS modes.

TLS is controlled through environment configuration:

```env
GRPC_USE_TLS=false
```

For local development, TLS can be disabled.

For TLS mode:

```env
GRPC_USE_TLS=true
GRPC_CERT_FILE=/certs/server.crt
GRPC_KEY_FILE=/certs/server.key
GRPC_CA_FILE=/certs/server.crt
```

The gRPC server loads its certificate and private key, while the Worker verifies the server certificate using the configured CA certificate.

Local certificate files are excluded from Git and should not be committed to the repository.

Both TLS and plaintext modes were tested with end-to-end task processing.

---

# Observability

The project includes logging, metrics, health checks, and distributed tracing.

## Structured Logging

Application and worker events use Go's `log/slog` with structured fields such as:

```text
request_id
task_id
worker_id
user_id
task_type
error
```

Structured logging makes it easier to correlate application events and investigate failures.

---

## Prometheus Metrics

The Worker exposes Prometheus-compatible metrics at:

```text
http://localhost:9090/metrics
```

Metrics include:

```text
tasks_processed_total
tasks_failed_total
task_retries_total
task_processing_duration_seconds
task_queue_size
```

Different Prometheus metric types are used depending on the data:

- Counter — processed tasks, failures, retries
- Histogram — processing duration
- Gauge — current Redis queue size

---

## Health Checks

```http
GET /health
```

The health endpoint checks application dependencies including:

- PostgreSQL
- Redis

If a required dependency is unavailable, the endpoint returns an unhealthy response.

Health checks were tested by stopping PostgreSQL and Redis containers and verifying that the API reports dependency failure.

---

## Distributed Tracing

OpenTelemetry is used for tracing application operations.

The project includes tracing for:

```text
HTTP API
TaskService
Worker
gRPC Client
gRPC Server
```

Gin is instrumented with `otelgin`, while gRPC uses OpenTelemetry client and server instrumentation.

Trace context is propagated between:

```text
Worker → gRPC Sender
```

so the client and server spans share the same Trace ID.

The current Redis queue contains only the task ID, so trace context is not propagated across:

```text
API → Redis → Worker
```

This is a known limitation. A more advanced implementation could store trace context together with the queue message.

The current project uses the OpenTelemetry stdout exporter to make traces visible during local development.

---

# Tech Stack

- Go
- Gin
- PostgreSQL
- Redis
- gRPC / Protocol Buffers
- Prometheus
- OpenTelemetry
- `log/slog`
- Docker
- Docker Compose
- Swagger / OpenAPI
- Goroutines
- Channels
- pgx
- go-redis
- wrk

Concepts and practices:

- REST API
- Worker Pool
- Repository Pattern
- Service Layer
- Dependency Injection
- Dependency Inversion
- Composition Root
- DTO separation
- Structured Logging
- Metrics
- Distributed Tracing
- Health Checks
- Observability
- Graceful Shutdown
- Retry Strategy
- Idempotency Protection
- Compensating Rollback
- Unit Testing
- Integration Testing
- Load Testing

---

# Project Structure

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
├── tracing/
└── worker/

loadtests/
└── create_task.lua

proto/
├── notification.proto
└── notificationpb/

migrations/

docs/

docker-compose.yml
```

`internal/app` acts as the application Composition Root and is responsible for creating and wiring application dependencies.

Interfaces are defined on the consumer side so that handlers, services, and workers depend on behavior rather than concrete infrastructure implementations.

---

# Getting Started

## Clone repository

```bash
git clone https://github.com/sainakuo/scalable-notification-system.git

cd scalable-notification-system
```

## Configure environment

Create `.env` based on `.env.example`.

Example local configuration:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sns_db

REDIS_ADDR=localhost:6379

GRPC_SENDER_ADDR=localhost:50051
GRPC_USE_TLS=false

GRPC_CERT_FILE=
GRPC_KEY_FILE=
GRPC_CA_FILE=

API_PORT=8080
```

## Run with Docker

```bash
docker compose up --build
```

The following services will start:

```text
API
Worker
gRPC Sender
PostgreSQL
Redis
Database migrations
```

---

# API Endpoints

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

# Create Task

## curl

```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "user_id": 1,
    "type": "email",
    "payload": "Welcome!"
}'
```

## PowerShell

```powershell
Invoke-RestMethod `
-Uri "http://localhost:8080/tasks" `
-Method POST `
-ContentType "application/json" `
-Body '{"user_id":1,"type":"email","payload":"Welcome!"}'
```

Example task lifecycle:

```text
POST /tasks
    |
    v
pending
    |
    v
Redis Queue
    |
    v
Worker
    |
    v
gRPC Sender
    |
    v
done
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

Allowed task statuses are enforced by a database constraint:

```text
pending
processing
done
failed
```

Retry count is also protected by a database constraint.

Indexes are created for:

```text
status
created_at
user_id
```

Database migrations are designed to be safely executed multiple times using `IF NOT EXISTS`.

Connect to PostgreSQL:

```bash
docker exec -it sns_postgres psql -U postgres -d sns_db
```

Example:

```sql
SELECT
    id,
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
- Duplicate task protection
- Queue publishing rollback

Run:

```bash
go test ./...
```

---

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

> The Worker should be stopped while Redis queue integration tests are running because both the test and Worker consume the same Redis queue.

After testing:

```powershell
Remove-Item Env:INTEGRATION_TEST
```

---

# Load Testing

The API was load-tested locally using `wrk` while running the complete system through Docker Compose.

The POST `/tasks` scenario performs the full task-creation path:

```text
HTTP Request
    |
    v
Gin API
    |
    v
TaskService
    |
    +------> PostgreSQL INSERT
    |
    +------> Redis LPUSH
```

Test script:

```text
loadtests/create_task.lua
```

Results:

| Threads | Connections | Duration | Requests/sec | Avg latency | Max latency |
|---:|---:|---:|---:|---:|---:|
| 1 | 1 | 2s | ~226 | 4.81 ms | 45.39 ms |
| 2 | 10 | 10s | ~1109 | 9.19 ms | 44.79 ms |
| 4 | 50 | 10s | ~911 | 26.83 ms | 99.39 ms |
| 4 | 100 | 10s | **~1312** | **48.35 ms** | 105.40 ms |

The highest measured throughput in this local Docker environment was approximately:

```text
1312 POST requests/second
```

After the load test, the Redis queue was empty and all tested tasks had reached the `done` state.

These results are **local benchmark results only** and should not be interpreted as production capacity. Performance depends on hardware, container configuration, database settings, network latency, worker count, and workload characteristics.

---

# Design Decisions and Trade-offs

This project intentionally keeps several reliability mechanisms understandable rather than trying to implement every distributed-systems pattern.

Current decisions include:

```text
PostgreSQL + Redis consistency
→ compensating rollback
→ production alternative: Transactional Outbox

Delivery semantics
→ at-least-once processing
→ idempotency protection in Redis
→ no strict exactly-once guarantee

Tracing through Redis
→ task ID only
→ API and Worker traces are separate
→ production alternative: propagate trace context in queue messages

GET /tasks
→ currently returns all tasks
→ production improvement: pagination
```

These limitations are intentionally documented rather than hidden.

---

# What This Project Demonstrates

This project demonstrates practical backend development experience with:

- Designing REST APIs in Go
- Layered backend architecture
- Service and Repository layers
- PostgreSQL persistence
- Redis queues
- Asynchronous processing
- Worker pools and concurrency
- Goroutines and channels
- Retry strategies
- Idempotency protection
- gRPC service communication
- TLS configuration
- Error propagation and domain errors
- Dependency inversion
- Composition Root
- Context propagation
- Graceful shutdown
- Structured application logging
- Prometheus metrics
- OpenTelemetry tracing
- Dependency health checks
- Database indexes and constraints
- Unit and integration testing
- Load testing
- Dockerized multi-service applications

---

# Roadmap

Completed:

```text
REST API and PostgreSQL persistence
Redis task queue
Worker Pool
gRPC notification sender
Retry strategy
Graceful Shutdown
Idempotency protection
Compensating rollback
Structured logging
Health checks
Prometheus metrics
OpenTelemetry tracing
Optional gRPC TLS
Database indexes and constraints
Unit tests
Integration tests
wrk load testing
Performance analysis
```

Possible future improvements:

```text
Transactional Outbox
Stronger atomic idempotency guarantees
Trace context propagation through Redis
Pagination for GET /tasks
Configurable worker pool size
Exponential backoff for retries
External OpenTelemetry collector / Jaeger
CI pipeline
```