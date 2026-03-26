# Subscription API

A RESTful microservice for managing user subscriptions and calculating their accumulated costs over time. Built with Go and the Echo framework, this service follows Clean Architecture principles to ensure separation of concerns, scalability, and maintainability.

## Quick Start

### 1. Configuration

Create a `.env` file in the root directory of the project. You can use the following template to configure your environment:

```env
# Application configuration
ENV=development
HTTP_PORT=8080
ROOT_DIR=./data

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD={YOUR_DB_PASSWORD}
DB_NAME=subscriptions_db
```

Alternatively, if you are using a `config.yaml` file, it should look like this:

```yaml
env: "development"

server:
  timeout: 15s
  idle_timeout: 60s
  host: "0.0.0.0"
  port: 8080
  server_mode: "development"

database:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "{YOUR_DB_PASSWORD}"
  name: "subscriptions_db"
```

### 2. Environment Setup

The project uses `make` commands to simplify the build and deployment processes. Run the following commands to get the environment up and running:

```bash
# Clone the repository
git clone <repository_url>
cd <project_directory>

# Download Go dependencies
go mod download

# Spin up the database and required infrastructure via Docker
make env-build

# Start the application locally
make env-up
```

*Note: If you do not have a Makefile configured yet, you can run the app directly using `go run cmd/app/main.go`.*

## API Reference

All application routes are prefixed with `/api/v1`. 

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/subscriptions` | Create a new subscription |
| `GET` | `/subscriptions` | Retrieve a paginated list of user subscriptions (`?user_id=...&limit=10&offset=0`) |
| `GET` | `/subscriptions/calculations` | Calculate total subscription cost over a period |
| `GET` | `/subscriptions/:id` | Get detailed information about a specific subscription |
| `PUT` | `/subscriptions/:id` | Update an existing subscription by ID |
| `DELETE`| `/subscriptions/:id` | Delete a subscription by ID |

### Cost Calculation Endpoint
The `/subscriptions/calculations` endpoint allows users to calculate how much they have spent on subscriptions.
**Query Parameters:**
* `user_id` (required): The UUID of the user.
* `start_date` (required): Calculation start date (format: `MM-YYYY`).
* `end_date` (optional): Calculation end date. Defaults to the current date if omitted.
* `service_name` (optional): Filter calculation for a specific service (e.g., "Netflix").

## Swagger Documentation

This API is fully documented using Swagger. Once the server is running, you can access the interactive API documentation via your browser:

**URL:** `http://localhost:8080/swagger/index.html`

To regenerate the Swagger documentation after updating the handler comments, run:

```bash
swag init -g cmd/app/main.go --parseDependency
```

## Tech Stack

* **Language:** Go 1.20+
* **Web Framework:** Echo (`github.com/labstack/echo/v4`)
* **Database:** PostgreSQL
* **Validation:** Go Playground Validator (`github.com/go-playground/validator/v10`)
* **Logging:** Standard `log/slog`
* **API Documentation:** Swaggo (`github.com/swaggo/echo-swagger`)

## Architecture

The project strictly adheres to Clean Architecture:
* **Handler Layer (`internal/handler`):** Parses HTTP requests, validates input using struct tags, and formats HTTP responses.
* **Usecase Layer (`internal/usecase`):** Contains the core business logic (e.g., cost calculation, date validation).
* **Repository Layer (`internal/database/postgres`):** Handles all data persistence and direct interaction with the PostgreSQL database.

## Graceful Shutdown

The application implements graceful shutdown procedures. Upon receiving an interrupt signal (`SIGINT` or `SIGTERM`), the server will stop accepting new requests, finish processing active requests within a 10-second timeout window, safely close the database connection pool, and then exit.
