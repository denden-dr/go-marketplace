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
- **DB Migrations**: Automated database versioning and migrations.

---

## 🛠 Tech Stack

- **Language**: [Go](https://go.dev/) (v1.25.8+)
- **Web Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: [PostgreSQL](https://www.postgresql.org/)
- **Database Tooling**: [pgxpool](https://github.com/jackc/pgx) (Driver/Pool), [golang-migrate](https://github.com/golang-migrate/migrate)
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
└── server/      # Router & Fiber initialization
```

This structure ensures that features are loosely coupled and easy to test or refactor independently.

---

## 🧪 Testing

To run the full test suite with verbose output:
```bash
make test
```
