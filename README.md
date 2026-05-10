# Go Marketplace

A high-performance e-commerce backend built with Go, featuring a **Rich Domain Model** and feature-based architecture.

## 🚀 Overview

**Go Marketplace** is a robust e-commerce API designed for scalability and maintainability. It follows clean architecture principles, encapsulating business logic within domain entities to ensure a decoupled and testable codebase.

### Key Features
- **Custom Authentication**: Local email/password registration with 2FA/Verification codes via MailerSend.
- **Rich Domain Model**: Business logic and state transitions encapsulated in domain entities.
- **Wallet & Escrow System**: Secure internal wallet with a transaction-safe escrow mechanism for order processing.
- **Merchant System**: Multi-vendor support with shop management and product catalogs.
- **Advanced Search**: Semantic and full-text search powered by PostgreSQL `pgvector` and `tsvector`.
- **Cart & Checkout**: Real-time cart management and atomic order processing.
- **Health Monitoring**: Real-time component health checks (Database, MailerSend, etc.).
- **Automated Migrations**: Database versioning and schema management with `golang-migrate`.

---

## 🛠 Tech Stack

- **Language**: [Go](https://go.dev/) (v1.25.8+)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Data Access**: [sqlx](https://github.com/jmoiron/sqlx) & [pgx](https://github.com/jackc/pgx)
- **Search**: [pgvector](https://github.com/pgvector/pgvector)
- **Authentication**: JWT (RS256/HS256) & [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Email**: [MailerSend SDK](https://github.com/mailersend/mailersend-go)
- **Testing**: [testify](https://github.com/stretchr/testify), [sqlmock](https://github.com/DATA-DOG/go-sqlmock), [mockery](https://github.com/vektra/mockery)

---

## ⚙️ Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (v1.25.8+)
- [PostgreSQL](https://www.postgresql.org/download/) with [pgvector](https://github.com/pgvector/pgvector) installed.
- [Docker](https://www.docker.com/) (for integration tests).

### Installation & Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/denden-dr/go-marketplace.git
   cd go-marketplace
   ```

2. **Configure Environment Variables**:
   Create a `.env` file based on `.env.example`:
   ```env
   PORT=3000
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASS=postgres
   DB_NAME=marketplace
   JWT_SECRET=your_super_secret_key
   MAILERSEND_API_KEY=mlsn.xxxxxx
   MAILERSEND_FROM_EMAIL=no-reply@yourdomain.com
   ```

3. **Install Dependencies**:
   ```bash
   make tidy
   ```

4. **Run Database Migrations**:
   ```bash
   make migrate-up
   ```

5. **Run the Application**:
   ```bash
   make run
   ```

---

## 🏗 Architecture

The project follows a **Feature-Based Architecture** with a **Rich Domain Model**. Each feature in `internal/core/` is isolated and self-contained.

### Core Layers
1. **Handler**: Fiber HTTP layer. Parses requests, validates DTOs, and calls the Service.
2. **Service**: Business orchestrator. Coordinates between repositories and domain entities.
3. **Domain Entity**: The heart of the application. Contains business rules, state validation, and transitions.
4. **Repository**: Data access layer. Handles SQL queries using `sqlx`.

### Directory Structure
```text
internal/
├── common/         # Shared utilities (Response wrappers, etc.)
├── core/           # Feature domains
│   └── [feature]/  # e.g., auth, wallet, order
│       ├── domain.go      # Business logic & Entities
│       ├── *_handler.go   # HTTP handlers
│       ├── *_service.go   # Service orchestrator
│       └── *_repo.go      # Data access
├── database/       # DB connection & migrations
├── domain/         # Shared structs & global errors
├── middleware/     # Fiber middlewares (Auth, Logger, etc.)
└── server/         # Router & Server setup
```

---

## 🧪 Testing

The project emphasizes a strong testing culture with high coverage of both unit and integration tests.

- **Unit Tests**: Test domain logic and service orchestration using mocks.
- **Integration Tests**: Test repository layer and database constraints using real PostgreSQL instances in Docker.

**Run all tests**:
```bash
make test
```

**Run integration tests (requires Docker)**:
```bash
make test-docker
```

---

## 🛡️ Security & Performance

- **Refresh Token Rotation**: Implements a "Token Family" pattern to detect and revoke reused refresh tokens.
- **Rate Limiting**: Applied to sensitive endpoints (Auth, Wallet) to prevent abuse.
- **Semantic Search**: Utilizes `pgvector` for efficient similarity searches, enabling features like "Recommended Products".

---

## 📖 Makefile Commands

| Command | Description |
| :--- | :--- |
| `make build` | Build the application binary. |
| `make run` | Run the application locally. |
| `make test` | Run all unit tests. |
| `make test-docker` | Run integration tests using Docker. |
| `make migrate-up` | Apply database migrations. |
| `make swagger` | Generate Swagger API documentation. |
| `make mock` | Regenerate mocks for testing. |

---

## 🏥 Health Monitoring

Monitor the status of your service and its dependencies:
- **Endpoint**: `GET /api/health`
- **Success**: `200 OK`
- **Failure**: `503 Service Unavailable`
