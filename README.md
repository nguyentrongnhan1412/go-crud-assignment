# Simple Product Management API

A RESTful API for managing products in storage, built with Go, Gin, and PostgreSQL. The project follows a layered architecture (handler → service → repository) with interface-based dependency injection and supports full CRUD operations with input validation.

## Tech Stack

- **Go 1.26** – backend language
- **Gin** – HTTP web framework
- **PostgreSQL 16** – relational database
- **pgx/v5** – PostgreSQL driver (`database/sql` via `jackc/pgx/v5/stdlib`)
- **Docker & Docker Compose** – containerized local and deployment setup

## Project Structure

```
├── config/                         # Environment-based configuration
│   └── config_test.go
├── internal/
│   ├── app/                        # Application bootstrap and wiring
│   ├── middleware/                 # HTTP middleware (request timeout)
│   ├── routes/                     # HTTP route definitions
│   ├── handlers/                   # HTTP request handlers
│   │   └── product_service.go      # Handler-layer service interface
│   ├── services/                   # Business logic and validation
│   │   └── product_repository.go   # Service-layer repository interface
│   ├── repositories/               # PostgreSQL data access
│   ├── models/                     # Domain models and DTOs
│   └── infrastructure/             # Database connection pool setup
├── migrations/                     # SQL schema and seed data
├── main.go                         # Entry point
├── Dockerfile                      # Multi-stage production image
├── docker-compose.yml              # Local and deployment orchestration
└── .env.example                    # Environment variable reference
```

## Architecture

The application is wired in `internal/app/app.go`:

1. **Infrastructure** – opens a PostgreSQL connection pool with configurable limits and a startup health ping
2. **Repository** – executes SQL queries (CRUD + keyword search) using request-scoped context
3. **Service** – validates input and delegates to the repository
4. **Handler** – parses HTTP requests, maps errors to status codes, returns JSON
5. **Middleware** – applies a per-request timeout via `context.WithTimeout` so slow queries do not hold goroutines indefinitely

Each layer depends on an interface defined in the layer above, which keeps handlers and services testable with mocks.

### Operational defaults

| Concern | Default | Configuration |
|---------|---------|---------------|
| Request timeout | 30s | `REQUEST_TIMEOUT` |
| Max open DB connections | 25 | `DB_MAX_OPEN_CONNS` |
| Max idle DB connections | 5 | `DB_MAX_IDLE_CONNS` |
| Connection max lifetime | 5m | `DB_CONN_MAX_LIFETIME` |

Tune the connection pool against PostgreSQL `max_connections` and the number of API instances:

```
DB_MAX_OPEN_CONNS <= (max_connections - reserved) / number_of_api_instances
```

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) and Docker Compose (recommended)
- PostgreSQL 16+ (if running without Docker)

## Quick Start with Docker

The fastest way to run the API and database together. Docker Compose starts PostgreSQL, applies the migration and seed data automatically, then starts the API once the database is healthy.

```bash
# Copy environment variables (optional)
cp .env.example .env

# Build and start all services
docker compose up --build

# Run in detached mode
docker compose up --build -d
```

The API will be available at `http://localhost:8080`.

Verify the stack is running:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

Stop services:

```bash
docker compose down

# Remove volumes (resets database)
docker compose down -v
```

When running with Docker Compose, `DB_HOST` is set to `db` automatically for the API container. Other settings such as `REQUEST_TIMEOUT` and pool variables use application defaults unless you add them to `.env`.

## Run Locally Without Docker

### 1. Start PostgreSQL

Create the database and run the migration (includes sample seed data):

```bash
psql -U postgres -c "CREATE DATABASE product_management;"
psql -U postgres -d product_management -f migrations/001_create_products_table.sql
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Adjust values in `.env` if your PostgreSQL credentials or timeout/pool settings differ.

### 3. Run the Application

```bash
go mod tidy
go run main.go
```

Server starts on port `8080` by default.

## Running Tests

Unit tests cover the config, middleware, service, handler, and repository layers:

```bash
go test ./...
```

Verbose output:

```bash
go test -v ./...
```

Repository tests use `go-sqlmock`; handler and service tests use mock implementations of the layer interfaces.

## API Endpoints

| Method | Endpoint              | Description                              |
|--------|-----------------------|------------------------------------------|
| GET    | `/health`             | Health check                             |
| POST   | `/products`           | Create a new product (returns `201`)     |
| GET    | `/products`           | List all products                        |
| GET    | `/products?keyword=`  | Search products by name or description   |
| GET    | `/products/:id`       | Get product by ID                        |
| PUT    | `/products/:id`       | Update a product                         |
| DELETE | `/products/:id`       | Delete a product                         |

### Product Response

```json
{
  "id": 1,
  "name": "Mechanical Keyboard",
  "description": "Wireless mechanical keyboard",
  "price": 120.50,
  "quantity": 10,
  "created_at": "2026-06-27T10:00:00Z",
  "updated_at": "2026-06-27T10:00:00Z"
}
```

Keyword search is case-insensitive and matches against both `name` and `description`.

## Example cURL Commands

**Create Product**

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mechanical Keyboard",
    "description": "Wireless mechanical keyboard",
    "price": 120.50,
    "quantity": 10
  }'
```

**Get Products**

```bash
curl http://localhost:8080/products
```

**Search Products**

```bash
curl "http://localhost:8080/products?keyword=keyboard"
```

**Get Product Detail**

```bash
curl http://localhost:8080/products/1
```

**Update Product**

```bash
curl -X PUT http://localhost:8080/products/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Keyboard",
    "description": "Updated description",
    "price": 135.00,
    "quantity": 15
  }'
```

**Delete Product**

```bash
curl -X DELETE http://localhost:8080/products/1
# {"message":"product deleted successfully"}
```

## Validation Rules

| Field      | Rules                                              |
|------------|----------------------------------------------------|
| `name`     | Required, minimum 3 characters (after trimming)    |
| `price`    | Required, must be greater than 0                   |
| `quantity` | Required, must be greater than or equal to 0       |

`description` is optional.

## Common Errors

| Status | Response                                              | Cause                          |
|--------|-------------------------------------------------------|--------------------------------|
| 400    | `{ "message": "invalid request body" }`               | Malformed JSON body            |
| 400    | `{ "message": "name is required" }`                   | Missing or empty product name  |
| 400    | `{ "message": "name must be at least 3 characters" }` | Name too short                 |
| 400    | `{ "message": "price must be greater than 0" }`       | Invalid price                  |
| 400    | `{ "message": "quantity must be greater than or equal to 0" }` | Invalid quantity |
| 400    | `{ "message": "invalid product id" }`                 | Non-numeric or invalid ID      |
| 404    | `{ "message": "product not found" }`                  | Product does not exist         |
| 500    | `{ "message": "internal server error" }`              | Unexpected server error        |
| 500    | `{ "message": "failed to fetch products" }`           | List query failed              |
| 504    | `{ "message": "request timeout" }`                    | Request exceeded `REQUEST_TIMEOUT` |

## Deployment

For production deployment with Docker Compose:

1. Set secure values for `DB_USER`, `DB_PASSWORD`, and other variables in `.env`.
2. Tune `REQUEST_TIMEOUT` and `DB_MAX_OPEN_CONNS` for your workload and PostgreSQL capacity.
3. Use `docker compose up --build -d` on your server.
4. Place a reverse proxy (e.g. Nginx) in front of the API for HTTPS if needed.
5. Consider using an external managed PostgreSQL service and updating `DB_HOST` accordingly.

The production Docker image is built in two stages (Go 1.26 Alpine builder, Alpine 3.20 runtime) and runs as the `nobody` user.

## Environment Variables

| Variable               | Default               | Description                                      |
|------------------------|-----------------------|--------------------------------------------------|
| `SERVER_PORT`          | `8080`                | HTTP server port                                 |
| `REQUEST_TIMEOUT`      | `30s`                 | Per-request timeout (Go duration, e.g. `15s`, `1m`) |
| `DB_HOST`              | `localhost`           | PostgreSQL host                                  |
| `DB_PORT`              | `5432`                | PostgreSQL port                                  |
| `DB_USER`              | `postgres`            | Database user                                    |
| `DB_PASSWORD`          | `postgres`            | Database password                                |
| `DB_NAME`              | `product_management`  | Database name                                    |
| `DB_SSLMODE`           | `disable`             | PostgreSQL SSL mode                              |
| `DB_MAX_OPEN_CONNS`    | `25`                  | Max concurrent connections per API instance      |
| `DB_MAX_IDLE_CONNS`    | `5`                   | Idle connections kept in the pool (≤ open conns) |
| `DB_CONN_MAX_LIFETIME` | `5m`                  | Max age before a connection is recycled        |

See `.env.example` for a ready-to-copy template with tuning notes.
