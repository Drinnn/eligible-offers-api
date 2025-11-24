# Design Document - Eligible Offers API

## 📐 Architecture Overview

This project follows **Clean Architecture** principles with clear separation of concerns across layers:

```
┌─────────────────────────────────────────┐
│         HTTP Layer (Handlers)           │  ← Entry point, HTTP concerns
├─────────────────────────────────────────┤
│        Use Cases (Business Logic)       │  ← Core business rules
├─────────────────────────────────────────┤
│      Repositories (Data Access)         │  ← Data persistence abstraction
├─────────────────────────────────────────┤
│         Entities (Domain Models)        │  ← Core domain objects
└─────────────────────────────────────────┘
```

### Layer Responsibilities

**1. Entities** (`internal/entities/`)

- Pure domain models (Offer, Transaction)
- No external dependencies
- Business-critical data structures

**2. Repositories** (`internal/repositories/`)

- Data persistence interfaces and implementations
- Thread-safe in-memory storage using `sync.RWMutex`
- Easy to swap implementations (in-memory → Postgres → etc)

**3. Use Cases** (`internal/use_cases/`)

- Business logic and rules
- Orchestrates repositories
- Independent of HTTP layer
- Example: Eligibility calculation logic lives here

**4. Handlers** (`internal/handlers/`)

- HTTP request/response handling
- Input validation (structural)
- Delegates business logic to use cases
- Error handling through middleware

**5. DTOs** (`internal/dtos/`)

- Data Transfer Objects for API contracts
- Validation rules using `go-playground/validator`
- Decoupled from internal entities

---

## 🗄️ Data Storage

### Current Implementation: In-Memory

**Structure:**

```go
type InMemoryOfferRepository struct {
    mu     sync.RWMutex
    offers map[string]*entities.Offer
}

type InMemoryTransactionRepository struct {
    mu           sync.RWMutex
    transactions map[string]*entities.Transaction
}
```

**Concurrency:**

- `RWMutex` allows multiple concurrent reads
- Single writer locks for data consistency
- Thread-safe for concurrent HTTP requests

**Trade-offs:**

| Aspect           | In-Memory          | Postgres              |
| ---------------- | ------------------ | --------------------- |
| Setup complexity | ✅ Zero            | ❌ Requires setup     |
| Performance      | ✅ Sub-ms          | ⚠️ Network latency    |
| Durability       | ❌ Lost on restart | ✅ Persistent         |
| Scalability      | ⚠️ Single instance | ✅ Horizontal scaling |
| Query capability | ⚠️ Full scan       | ✅ Indexed queries    |

**Why In-Memory is Sufficient:**

- Requirements only specify data must survive "while the server is running"
- Test scope prioritizes correctness over production features
- Easy to swap later (see Migration Path below)

---

## 🧮 Eligibility Calculation Logic

### Algorithm (in `GetEligibleOffersUseCase`)

```
FOR each offer in all offers:
  1. Filter: Is offer ACTIVE?
     - offer.Active == true
     - offer.StartsAt <= now <= offer.EndsAt

  2. IF active, count matching transactions:
     - lookback_start = now - offer.LookbackDays

     FOR each user transaction:
       - Skip if transaction.ApprovedAt < lookback_start
       - Skip if transaction.ApprovedAt > now

       - Match if:
         * transaction.MerchantID == offer.MerchantID OR
         * transaction.MCC ∈ offer.MCCWhitelist

       - Increment count

  3. IF count >= offer.MinTxnCount:
     - User is ELIGIBLE
     - Add to result with reason
```

### Complexity Analysis

**Time Complexity:** O(O × T)

- O = number of offers
- T = number of user transactions

**Optimizations Considered (but not implemented):**

- Index transactions by merchant_id and MCC → O(O) lookup
- Pre-filter active offers in repository
- **Decision:** Keep simple for 3-hour scope; premature optimization avoided

**For Production Scale:**

- Use SQL with indexes: `WHERE merchant_id = ? OR mcc IN (?)`
- Materialize eligibility results (cache)
- Add pagination to `/eligible-offers` endpoint

---

## 🔀 Request Flow Example

### GET /users/{user_id}/eligible-offers?now=2025-11-23T10:00:00Z

```
1. HTTP Request
   ↓
2. Handler (GetEligibleOffersHandler)
   - Extract user_id from path param (chi.URLParam)
   - Parse optional 'now' query param (RFC3339)
   - Default to time.Now() if not provided
   ↓
3. Use Case (GetEligibleOffersUseCase)
   - Fetch all offers from repository
   - Filter active offers (business logic)
   - Fetch user transactions from repository
   - Calculate eligibility for each offer
   - Build response DTOs
   ↓
4. Handler
   - Return 200 OK with JSON response
   ↓
5. Middleware (ErrorHandler)
   - Catches errors and formats responses
```

---

## ✅ Validation Strategy

### Two-Layer Validation

**1. Structural Validation (in DTOs)**

```go
type UpsertOfferRequest struct {
    MinTxnCount  int `validate:"required,gt=0"`
    MCCWhitelist []string `validate:"required,dive,len=4,numeric"`
}
```

- Field presence, types, formats
- Using `go-playground/validator`
- Returns 400 Bad Request with field-level errors

**2. Business Rules (in Use Cases)**

```go
if !request.StartsAt.Before(request.EndsAt) {
    return customErrors.NewBadRequestError("starts_at must be before ends_at", nil)
}
```

- Domain-specific rules
- Relationships between fields
- Context-dependent validation

---

## 🧪 Testing Strategy

### Unit Tests (`internal/use_cases/*_test.go`)

- **Focus:** Business logic in isolation
- **Coverage:** 6 test cases for eligibility logic
- **Examples:**
  - User qualifies (happy path)
  - Not enough transactions
  - Offer inactive
  - Offer expired
  - Transactions outside lookback window
  - Match by MCC (not merchant_id)

### Integration Tests (`tests/integration/*_test.go`)

- **Focus:** End-to-end API flows
- **Approach:** `httptest.Server` with real repositories
- **Coverage:** 3 scenarios testing full request/response cycle
- **Examples:**
  - Create offer → Ingest txns → Check eligibility (eligible)
  - Create offer → Ingest txns → Check eligibility (not eligible)
  - Match by MCC across different merchants

---

## 🎯 Design Decisions & Trade-offs

### Decision: In-Memory Storage

**Rationale:** Requirements only need data to survive during server runtime
**Trade-off:** No persistence across restarts, but meets scope
**Future:** Easy to swap for Postgres via interface

### Decision: Business Logic in Use Case

**Rationale:** Keep repository "dumb" (CRUD only), logic testable
**Trade-off:** Repository can't optimize queries (e.g., SQL WHERE)
**Future:** Move filtering to repository when using SQL

### Decision: No Pagination

**Rationale:** Out of scope for 3-hour exercise
**Trade-off:** Could return many offers if user is super-eligible
**Future:** Add `?limit=10&offset=20` or cursor-based pagination

### Decision: Thread-safe Maps with RWMutex

**Rationale:** Allows concurrent reads (most operations)
**Trade-off:** Still single-instance limitation
**Future:** Database handles concurrency natively

---

## 📊 Performance Characteristics

### Current (In-Memory)

- **Offer Upsert:** O(1) - map insert
- **Transaction Insert:** O(N) - N = batch size, checking duplicates
- **Get Eligible Offers:** O(O × T) - O offers, T user transactions

### With Postgres (Future)

- **Offer Upsert:** O(log N) - B-tree index
- **Transaction Insert:** O(N log M) - batch insert with index
- **Get Eligible Offers:** O(O + T × log M) - indexed queries

### Bottleneck Analysis

**Current Bottleneck:** Eligibility calculation (full scan)
**Solution:** Index transactions by merchant_id + MCC in SQL

---

## 🔒 Security Considerations

### Implemented

✅ Input validation (structural)
✅ Thread-safe concurrent access
✅ No SQL injection (not using SQL)

### Not Implemented (Out of Scope)

❌ Authentication/Authorization
❌ Rate limiting
❌ Input sanitization (XSS, etc.)
❌ HTTPS/TLS
❌ Audit logging

---

## 🏁 Summary

This design prioritizes:

1. ✅ **Correctness** - Eligibility logic is accurate and tested
2. ✅ **Simplicity** - Easy to understand and maintain
3. ✅ **Testability** - Comprehensive unit + integration tests
4. ✅ **Extensibility** - Clean interfaces allow easy storage swap

The architecture is production-ready in structure, with intentional scope limitations appropriate for a 3-hour technical assessment.
