# Go Shop Yourself

A robust e-commerce backend built with Go, featuring a clean, feature-based architecture.

## 🚀 Overview

**Go Shop Yourself** is a comprehensive e-commerce API service designed for scalability and maintainability. It provides a full set of features for users, merchants, products, and order management, all powered by High-performance technologies like Go and Fiber.

### Key Features
- **Authentication**: JWT-based auth with refresh token family support for secure sessions.
- **User Management**: Profile management and secure password handling.
- **Merchant System**: Register as a merchant to sell products and manage shop settings.
- **Product Catalog**: Advanced product management including categories and stock tracking.
- **Shopping Cart**: Real-time cart management with price calculations.
- **Order Processing**: Complete checkout flow, order status tracking, and history.
- **Wallet System**: Internal wallet for users to manage balances and pay for orders.
- **Product Search**: Robust full-text and fuzzy search powered by PostgreSQL `pg_trgm` and `tsvector`.
- **Health Monitoring**: Real-time monitoring of application, database, and OpenSearch connectivity.
- **DB Migrations**: Automated database versioning and migrations.

---

## 🛠 Tech Stack

- **Language**: [Go](https://go.dev/) (v1.25.8+)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Database Tooling**: [pgxpool](https://github.com/jackc/pgx) (Driver/Pool), [golang-migrate](https://github.com/golang-migrate/migrate)
- **Search Engine**: [OpenSearch](https://opensearch.org/)
- **JSON Web Tokens**: [golang-jwt](https://github.com/golang-jwt/jwt)
- **Unit Testing**: [testify](https://github.com/stretchr/testify), [pgxmock](https://github.com/pashagolub/pgxmock), [mockery](https://github.com/vektra/mockery)
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
   git clone https://github.com/yourusername/go-shop-yourself.git
   cd go-shop-yourself
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

### 🔐 Firebase Auth Emulator

For local development, the application is configured to use the **Firebase Auth Emulator**. This allows you to test social login and token verification without real Firebase credentials.

1.  **Install Firebase CLI**:
    ```bash
    npm install -g firebase-tools
    ```

2.  **Start the Auth Emulator**:
    ```bash
    firebase emulators:start --only auth
    ```
    The emulator will run at `localhost:9099` (Auth) and `localhost:4000` (UI Dashboard) by default.

3.  **Application Config**:
    When `APP_ENV=development`, the application automatically connects to the emulator. You can customize the emulator host and project ID via environment variables (see below).

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
├── health/      # Application health checks (DB, OS)
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

The application provides a public health check endpoint to monitor the status of the service and its dependencies (Database and OpenSearch).

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
      "opensearch": "up"
    }
  }
}
```

---

## 🔍 OpenSearch Integration

The project integrates with **OpenSearch** for advanced search capabilities and real-time indexing.

### Purpose
- **Health Monitoring**: Connectivity checks are performed as part of the system health status.
*Next Steps*: OpenSearch will be used to power product search features, providing fast, full-text search and filtering.

### Configuration
Configure OpenSearch using the following environment variables in your `.env` file:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `OPENSEARCH_HOST` | Hostname of the OpenSearch cluster | `localhost` |
| `OPENSEARCH_PORT` | Port for OpenSearch service | `9200` |
| `OPENSEARCH_USER` | Username for authentication | `admin` |
| `OPENSEARCH_PASSWORD` | Password for authentication | `admin` |

### Connection logic
The application uses the `opensearch-go/v3` client. Connections are initialized in `main.go` and verified during startup. If OpenSearch is unavailable, the application will log a warning but continue to run (depending on feature availability).
