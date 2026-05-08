# Go Marketplace

A robust e-commerce backend built with Go, featuring a clean, feature-based architecture.

## 🚀 Overview

**Go Marketplace** is a comprehensive e-commerce API service designed for scalability and maintainability. It provides a full set of features for users, merchants, products, and order management, all powered by High-performance technologies like Go and Fiber.

### Key Features
- **Authentication**: JWT-based auth with refresh token family support for secure sessions.
- **User Management**: Profile management and secure password handling.
- **Merchant System**: Register as a merchant to sell products and manage shop settings.
- **Product Catalog**: Advanced product management including categories and stock tracking.
- **Shopping Cart**: Real-time cart management with price calculations.
- **Order Processing**: Complete checkout flow, order status tracking, and history.
- **Wallet System**: Internal wallet for users to manage balances and pay for orders.
- **Health Monitoring**: Real-time monitoring of application, database, and Supabase connectivity.
- **Product Search**: Robust semantic and full-text search powered by PostgreSQL `pg_trgm`, `tsvector`, and `pgvector`.
- **DB Migrations**: Automated database versioning and migrations.

---

## 🛠 Tech Stack

- **Language**: [Go](https://go.dev/) (v1.25.8+)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Database Tooling**: [sqlx](https://github.com/jmoiron/sqlx) (abstraction), [pgx](https://github.com/jackc/pgx) (stdlib driver), [golang-migrate](https://github.com/golang-migrate/migrate)
- **Search Engine**: [PostgreSQL pgvector](https://github.com/pgvector/pgvector)
- **JSON Web Tokens**: [golang-jwt](https://github.com/golang-jwt/jwt)
- **Unit Testing**: [testify](https://github.com/stretchr/testify), [sqlmock](https://github.com/DATA-DOG/go-sqlmock), [mockery](https://github.com/vektra/mockery)
- **Utilities**: [godotenv](https://github.com/joho/godotenv), [shopspring/decimal](https://github.com/shopspring/decimal)

---

## ⚙️ Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) installed.
- [PostgreSQL](https://www.postgresql.org/download/) running.
- `golang-migrate` installed (for database migrations).

### Installation & Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/go-marketplace.git
   cd go-marketplace
   ```

2. **Configure Environment Variables**:
   Create a `.env` file in the root directory and configure it based on your setup:
   ```env
   PORT=your_app_port
   DB_HOST=your_db_host
   DB_PORT=your_db_port
   DB_USER=your_db_user
   DB_PASS=your_db_password
   DB_NAME=your_dbname
   JWT_SECRET=your_jwt_secret_key_here
   APP_ENV=development
   FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
   FIREBASE_PROJECT_ID=fb-go-commerce-auth
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

### 🔐 Supabase Auth
 
 The application uses **Supabase Auth** for social login (Google, Facebook, Apple, Twitter). 
 
 1.  **Client-Side Flow**: Supabase handles the OAuth flow on the frontend and issues a JWT access token.
 2.  **Backend Verification**: The backend verifies the JWT using the `SUPABASE_JWT_SECRET` (HS256).
 3.  **Endpoint**: Use `POST /api/auth/social` with the `access_token` from Supabase.

---

---

## 🛡️ Security

### Rate Limiting
To prevent brute-force attacks and abuse, the following rate limits are applied to authentication endpoints (`/api/auth/*`):
- **Limit**: 10 requests per minute per IP address.
- **Scope**: Includes `/login`, `/register`, `/social`, and `/refresh`.
- **Response**: Exceeding the limit results in a `429 Too Many Requests` error.

### Input Validation
The API enforces strict validation on all inputs. For example, Social access tokens are capped at 5KB to prevent oversized payload attacks.

---

## 📖 Makefile Commands

The project includes a `Makefile` for common development tasks:

| Command | Description |
| :--- | :--- |
| `make build` | Compile the application into a binary. |
| `make run` | Run the application directly. |
| `make test` | Run all unit tests. |
| `make mock` | Generate mocks using mockery. |
| `make migrate-up` | Apply all pending database migrations. |
| `make migrate-down` | Rollback the last database migration. |
| `make fmt` | Format the Go source code. |
| `make tidy` | Tidy up Go modules. |
| `make swagger` | Generate Swagger API documentation. |
| `make clean` | Remove the compiled binary. |

---

## 🏗 Architecture

The project follows a **Feature-Based Architecture**, where each domain (e.g., `auth`, `order`, `product`) is encapsulated in its own package under `internal/`. 

```text
internal/
├── auth/        # Authentication & Refresh Tokens
├── cart/        # Shopping Cart logic
├── database/    # Migrations & Connection setup
├── domain/      # Core business entities
├── merchant/    # Merchant & Shop management
├── order/       # Ordering & Checkout flow
├── product/     # Product catalog
├── user/        # User accounts & profiles
├── wallet/      # Digital balance & transactions
├── health/      # Application health checks (DB, Supabase)
└── server/      # Router & Fiber initialization
```

This structure ensures that features are loosely coupled and easy to test or refactor independently.

---

## 🧪 Testing

To run the full test suite with verbose output:
```bash
make test
```

---

## 🏥 Health Monitoring

The application provides a public health check endpoint to monitor the status of the service and its dependencies (Database and Supabase).

- **Endpoint**: `GET /api/health`
- **Response Format**: JSON
- **Success Response**: `200 OK`
- **Error Response**: `503 Service Unavailable` (if any component is down)

**Example Request**:
```bash
curl http://localhost:3000/api/health
```

**Example Response**:
```json
{
  "message": "application is healthy",
  "status": 200,
  "data": {
    "components": {
      "database": "up",
      "supabase": "configured"
    }
  }
}
```

---

## 🔍 Vector Search Integration

The project integrates with **pgvector** for advanced semantic search capabilities.

### Purpose
- **Semantic Search**: Leverages vector embeddings to provide more relevant search results beyond keyword matching.
- **Efficiency**: Integrated directly into PostgreSQL, reducing infrastructure complexity.

### Next Steps
The application will use `pgvector-go` to manage embeddings and perform similarity searches.
