# Go Marketplace

A high-performance e-commerce backend built with Go, featuring a **Rich Domain Model** and feature-based architecture.

## 🚀 Overview

**Go Marketplace** is a robust e-commerce API designed for scalability and maintainability. It follows clean architecture principles, encapsulating business logic within domain entities to ensure a decoupled and testable codebase.

### Key Features
- **Authentication**: Local email/password registration and Google OAuth2 integration with secure account linking.
- **Rich Domain Model**: Business logic and state transitions encapsulated in domain entities.
- **Wallet & Escrow System**: Secure internal wallet with a transaction-safe escrow mechanism. Funds are held in `pending_balance` until order delivery confirmation.
- **Unified Payment Interface**: Integrated support for internal wallet payments and external provider webhooks (Midtrans).
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
- **Authentication**: JWT (RS256/HS256) & Google OAuth2
- **Email**: [MailerSend SDK](https://github.com/mailersend/mailersend-go)
- **Testing**: [testify](https://github.com/stretchr/testify), [sqlmock](https://github.com/DATA-DOG/go-sqlmock), [mockery](https://github.com/vektra/mockery), [testcontainers-go](https://github.com/testcontainers/testcontainers-go)

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
   GOOGLE_CLIENT_ID=your_id
   GOOGLE_CLIENT_SECRET=your_secret
   GOOGLE_REDIRECT_URL=http://localhost:3000/api/auth/google/callback
   GOOGLE_LOGIN_REDIRECT_URL=http://localhost:5173/login-success
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
│   └── [feature]/  # e.g., auth, wallet, order, payment
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
- **Integration Tests**: Test repository and handler layers using **Testcontainers** for programmatic lifecycle management of PostgreSQL.

**Run all tests**:
```bash
make test
```

**Run integration tests (requires Docker)**:
```bash
make test-integration
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
| `make test-integration` | Run integration tests using Testcontainers. |
| `make migrate-up` | Apply database migrations. |
| `make swagger` | Generate Swagger API documentation. |
| `make mock` | Regenerate mocks for testing. |

---

## 🏥 Health Monitoring

Monitor the status of your service and its dependencies:
- **Endpoint**: `GET /api/health`
- **Success**: `200 OK`
- **Failure**: `503 Service Unavailable`
