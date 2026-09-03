# Rebutin Backend Starter - Golang Clean Architecture

A Golang RESTful API starter project built with **Uncle Bob's Clean Architecture** and a modern tech stack:

- **Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database**: [PostgreSQL](https://www.postgresql.org/) with `lib/pq` driver
- **Query Builder / Data Access**: [sqlx](https://github.com/jmoiron/sqlx) (Extends standard `database/sql`)
- **Logger**: [Logrus](https://github.com/sirupsen/logrus)
- **Validator**: [go-playground/validator v10](https://github.com/go-playground/validator)
- **UUID**: [google/uuid](https://github.com/google/uuid)
- **Environment Management**: [godotenv](https://github.com/joho/godotenv)

---

## Project Structure (Clean Architecture)

```text
.
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
├── config/
│   └── config.go                    # Environment & configuration loader
├── internal/
│   ├── app/
│   │   └── app.go                   # Core application setup & graceful shutdown
│   ├── domain/
│   │   ├── errors.go                # Custom domain errors
│   │   └── user.go                  # Entities, DTOs, Repository & Usecase Interfaces
│   ├── repository/
│   │   └── postgres/
│   │       └── user_repository.go   # SQLX Postgres implementation of UserRepository
│   ├── usecase/
│   │   └── user_usecase.go          # Business logic implementation
│   └── delivery/
│       └── http/
│           ├── router.go            # Gin router setup & route declarations
│           ├── handler/
│           │   └── user_handler.go  # HTTP Handlers
│           ├── middleware/
│           │   ├── logger.go        # Logrus request logging middleware
│           │   └── recovery.go      # Panic recovery middleware
│           └── response/
│               └── response.go      # Standard HTTP JSON response envelope
├── pkg/
│   ├── database/
│   │   └── postgres.go              # Database connection pool setup
│   ├── logger/
│   │   └── logger.go                # Logrus initialization
│   └── validator/
│       └── validator.go             # Custom JSON validator wrapper
├── .env.example
├── .env
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```
---

## Getting Started

### 1. Running Locally

```bash
# Download dependencies
go mod download

# Ensure PostgreSQL is running and configure the .env file
cp .env.example .env

# Run the application
make run
# or
go run ./cmd/api
```

### 2. Running via Docker Compose

```bash
# Start PostgreSQL & API Server
make docker-up

# Stop containers
make docker-down
```
---

## API Endpoints

### Health Check
- `GET /health`

### Users API (`/api/v1/users`)
- `POST /api/v1/users` - Create User
- `GET /api/v1/users` - List Users (Supports `?page=1&limit=10`)
- `GET /api/v1/users/:id` - Get User by UUID
- `PUT /api/v1/users/:id` - Update User
- `DELETE /api/v1/users/:id` - Delete User

---
