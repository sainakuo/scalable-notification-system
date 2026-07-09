#  Scalable Notification System

A scalable backend system for asynchronous task processing built with Go.

This project demonstrates how to build a production-like backend using Go, REST API, PostgreSQL, Redis, gRPC, Docker, goroutines, channels, and worker pools.

---

#  Features

- REST API with Gin
- PostgreSQL storage
- Redis queue
- Background workers
- Worker Pool using goroutines and channels
- Retry mechanism for failed tasks
- gRPC communication between services
- Docker Compose deployment
- Swagger/OpenAPI documentation
- Graceful Shutdown
- Configuration via environment variables
- Database migrations

---

#  Architecture

```text
                        +----------------+
                        |    Client      |
                        +-------+--------+
                                |
                           HTTP REST API
                                |
                                v
                     +--------------------+
                     |     Gin API        |
                     +---------+----------+
                               |
                 Save task to PostgreSQL
                               |
                               v
                     +--------------------+
                     |    PostgreSQL      |
                     +---------+----------+
                               |
                    Push task ID to Redis
                               |
                               v
                     +--------------------+
                     |    Redis Queue     |
                     +---------+----------+
                               |
                               v
                    Worker Pool (5 goroutines)
               +--------+--------+--------+
               |        |        |        |
             Worker1 Worker2 Worker3 ...
               |        |        |
               +--------+--------+
                        |
                  gRPC Notification
                        |
                        v
              +----------------------+
              | gRPC Sender Service  |
              +----------------------+
```

---

#  Tech Stack

- Go
- Gin
- PostgreSQL
- Redis
- gRPC
- Docker
- Docker Compose
- Swagger
- Goroutines
- Channels

---

#  Project Structure

```text
cmd/
├── api/
├── worker/
└── grpc-sender/

internal/
├── config/
├── handler/
├── model/
├── queue/
└── repository/

proto/
├── notification.proto
└── notificationpb/

migrations/

docs/

docker-compose.yml
```

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

The following services will start automatically:

- API
- Worker
- gRPC Sender
- PostgreSQL
- Redis

---

#  API Endpoints

| Method | Endpoint | Description |
|---------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/tasks` | Create task |
| GET | `/tasks` | Get all tasks |
| GET | `/tasks/{id}` | Get task by ID |

Swagger:

```
http://localhost:8080/swagger/index.html
```

---

#  Create Task

### curl

```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "user_id":1,
    "type":"email",
    "payload":"Welcome!"
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

#  Retry Mechanism

If task processing fails:

```text
pending
    ↓
processing
    ↓
error
    ↓
retry_count + 1
    ↓
Redis Queue
```

After reaching the retry limit:

```text
failed
```

---

#  Check Tasks

Connect to PostgreSQL

```bash
docker exec -it sns_postgres psql -U postgres -d sns_db
```

Example query

```sql
SELECT id,
       type,
       status,
       retry_count
FROM tasks
ORDER BY id DESC;
```

---

#  What this project demonstrates

This project demonstrates practical experience with:

- Building REST APIs
- Repository pattern
- PostgreSQL integration
- Redis queues
- Background processing
- Worker Pool
- Goroutines
- Channels
- gRPC
- Docker
- Graceful Shutdown
- Retry mechanisms
- Clean Architecture principles

---