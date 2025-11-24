# Eligible Offers API

A Go-based HTTP service that determines which offers a user is eligible for based on their transaction history.

## Development Notes

This project was intentionally scoped to **~3 hours** of development time to focus on solving the core problem without over-engineering. You can follow the development progress through the commit history, which shows the iterative approach taken to build this solution.

---

## Running the Server

### Prerequisites
- Go 1.22 or higher

### Start the server

```bash
# Using Go directly
go run cmd/api/main.go

# Or using Make
make run
```

The server will start on `http://localhost:8080`

---

## Running Tests

### Unit Tests

```bash
# Run unit tests
make test

# Or using Go directly
go test ./internal/... -v
```

### Integration Tests

```bash
# Run integration tests
make test-integration

# Or using Go directly
go test ./tests/integration/... -v
```

### All Tests

```bash
# Run all tests (unit + integration)
make test-all

# Or using Go directly
go test ./... -v
```

---

## Storage Implementation

This implementation uses **in-memory storage** (Go maps) for simplicity and to meet the requirement:

> "It just needs to survive while the server is running so that eligibility checks work across requests."

### Why In-Memory?

- **Simple**: No external dependencies or database setup required
- **Fast**: Sub-millisecond query performance
- **Sufficient**: Meets the requirement of persisting data during server runtime
- **Thread-safe**: Uses `sync.RWMutex` for concurrent access

### Easy to Replace

The architecture follows **Clean Architecture** principles with clearly defined repository interfaces. Swapping the in-memory implementation for Postgres, SQLite, or any other database is straightforward:

1. Implement the `OfferRepository` and `TransactionRepository` interfaces
2. Update dependency injection in `main.go`
3. No changes needed in use cases, handlers, or entities

For more details, see [DESIGN.md](DESIGN.md)

---

## API Endpoints

### 1. Create/Update Offer
```bash
POST /offers
Content-Type: application/json

{
  "merchant_id": "uuid",
  "mcc_whitelist": ["5812", "5814"],
  "active": true,
  "min_txn_count": 3,
  "lookback_days": 30,
  "starts_at": "2025-10-01T00:00:00Z",
  "ends_at": "2025-10-31T23:59:59Z"
}
```

### 2. Ingest Transactions
```bash
POST /transactions
Content-Type: application/json

{
  "transactions": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "merchant_id": "uuid",
      "mcc": "5812",
      "amount_cents": 1250,
      "approved_at": "2025-10-20T12:34:56Z"
    }
  ]
}
```

### 3. Get Eligible Offers
```bash
GET /users/{user_id}/eligible-offers?now=2025-10-21T10:00:00Z

# The 'now' query parameter is optional (defaults to server time)
```

---

## Example Usage

```bash
# 1. Create an offer
curl -X POST http://localhost:8080/offers \
  -H "Content-Type: application/json" \
  -d '{
    "merchant_id": "merchant-123",
    "mcc_whitelist": ["5812"],
    "active": true,
    "min_txn_count": 2,
    "lookback_days": 30,
    "starts_at": "2025-01-01T00:00:00Z",
    "ends_at": "2025-12-31T23:59:59Z"
  }'

# 2. Ingest transactions
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "transactions": [
      {
        "id": "txn-1",
        "user_id": "user-456",
        "merchant_id": "merchant-123",
        "mcc": "5812",
        "amount_cents": 1000,
        "approved_at": "2025-11-20T12:00:00Z"
      },
      {
        "id": "txn-2",
        "user_id": "user-456",
        "merchant_id": "merchant-123",
        "mcc": "5812",
        "amount_cents": 2000,
        "approved_at": "2025-11-21T12:00:00Z"
      }
    ]
  }'

# 3. Check eligibility
curl "http://localhost:8080/users/user-456/eligible-offers?now=2025-11-23T10:00:00Z"
```

---

## Project Structure

```
eligible-offers-api/
├── cmd/api/              # Application entry point
├── internal/
│   ├── dtos/            # Request/Response data transfer objects
│   ├── entities/        # Domain entities
│   ├── handlers/        # HTTP handlers
│   ├── helpers/         # Utility functions
│   ├── middlewares/     # HTTP middlewares
│   ├── repositories/    # Data access layer
│   ├── use_cases/       # Business logic
│   └── errors/          # Custom error types
└── tests/
    └── integration/     # Integration tests
```

---

## Architecture & Design

For detailed information about:
- Architecture decisions
- Eligibility calculation logic
- Trade-offs and future improvements

See **[DESIGN.md](DESIGN.md)**

---

## Available Make Commands

```bash
make help              # Show all available commands
make run               # Run the server locally
make test              # Run unit tests
make test-integration  # Run integration tests
make test-all          # Run all tests
make clean             # Clean build artifacts
```

---

## What Was Intentionally Skipped

To keep the implementation within the 3-hour scope:
- Database persistence across server restarts (not required)
- API authentication/authorization
- Request pagination
- Graceful shutdown handling
- Metrics and observability
- API documentation (Swagger/OpenAPI)
- Rate limiting
- Input sanitization beyond basic validation

These would be priorities for a production system but were outside the scope of this exercise.
