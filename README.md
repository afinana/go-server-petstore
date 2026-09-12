# Go Petstore Backend Service

A high-performance Go REST API service backed by MongoDB, designed to operate behind an API Gateway such as **KrakenD**.

---

## Architecture Overview

```
[ Client ] 
    │
    │  Authorization: Bearer <JWT>
    ▼
┌────────────────────────────────────────────────────────┐
│ KrakenD API Gateway                                    │
│  - Edge TLS termination                                │
│  - JWT validation (Signature, Expiration, Issuer)      │
│  - Rate limiting & Request throttling                  │
│  - Claim extraction & header propagation               │
└────────────────────────────────────────────────────────┘
    │
    │  X-User-Id: <user-id>
    │  X-User-Roles: <roles>
    │  X-User-Email: <email>
    │  X-Gateway-Secret: <secret>
    ▼
┌────────────────────────────────────────────────────────┐
│ Go Petstore Backend Service                            │
│  - Gateway secret verification (defense-in-depth)      │
│  - Fine-grained Authorization (ABAC / Object Ownership)│
│  - Deep Input Validation (length, types, schema)       │
│  - Business Logic Limits (Max 3 adoptions per user)    │
│  - MongoDB Persistence                                 │
└────────────────────────────────────────────────────────┘
```

In this microservice architecture:
1. **KrakenD API Gateway** terminates external traffic, validates JWT signatures and expiration against the Identity Provider (e.g. Google OAuth2, Keycloak, Auth0, Okta), and propagates verified user identity headers downstream.
2. **Go Petstore Backend** does not duplicate cryptographic JWT verification. Instead, it ingests verified identity headers, verifies the internal gateway trust boundary, and enforces database-aware **Fine-Grained Authorization (ABAC)**, **Input Validation**, and **Business Logic Limits**.

---

## Security & Authorization Model

### 1. Gateway Identity Propagation & Authentication
The service extracts caller identity from incoming gateway headers:
- `X-User-Id` (or `X-Forwarded-User`): Unique user identifier / subject (`sub`).
- `X-User-Roles` (or `X-User-Role`): Comma-separated list of user roles (e.g., `user`, `admin`).
- `X-User-Email`: Authenticated user email.
- `X-Gateway-Secret`: Optional shared secret header matching `GATEWAY_SECRET` to prevent direct access bypassing the gateway.

### 2. Fine-Grained Authorization (ABAC / Resource-Level Ownership)
Because KrakenD is stateless and unaware of database state, resource ownership is enforced within the service:
- **Pet Ownership**:
  - Every pet created via `POST /v2/pet` has its `ownerId` set to the caller's `X-User-Id`.
  - Updating (`PUT /v2/pet`) or deleting (`DELETE /v2/pet/{petId}`) a pet requires the caller to be either:
    - The registered **owner** of the pet (`ownerId == user.ID`), OR
    - An **administrator** (`X-User-Roles: admin`).
  - Unauthorized modifications are rejected with `403 Forbidden`.
- **Order Ownership**:
  - Every placed order records `userId`.
  - Fetching an order (`GET /v2/store/order/{orderId}`) or cancelling an order (`DELETE /v2/store/order/{orderId}`) is restricted to the order owner or administrators.
  - Non-admin calls to `GET /v2/store/inventory` return only the authenticated user's orders.

### 3. Input Validation & Injection Prevention
All incoming requests are strictly validated prior to database interaction:
- **Payload Size Limiter**: Request bodies are capped at 1 MB (`http.MaxBytesReader`) to mitigate payload Denial-of-Service attacks.
- **Pet Validation**:
  - `name`: Required, trimmed, 1–100 characters, no ASCII control characters.
  - `status`: Restricted to allowed enum values (`available`, `pending`, `sold`).
  - `photoUrls`: Maximum 20 URLs, max 2048 characters each, strictly validated `http://` or `https://` schemes (preventing `javascript:` URI attacks).
  - `tags` / `category`: Bounded string lengths (max 50 chars).
- **Order Validation**:
  - `petId`: Must be a positive integer.
  - `quantity`: Must be between 1 and 3 (within adoption limits).
  - `status`: Restricted to `placed`, `approved`, `delivered`, `cancelled`.
- **Path Parameter Validation**:
  - Path parameters (`petId`, `orderId`) must be valid positive numeric integers or valid 24-character hexadecimal MongoDB ObjectIDs.
- **NoSQL Injection Protection**:
  - Query filters (`FindByStatus`, `FindByTags`) use typed BSON operators with the `$in` operator and sanitized inputs, eliminating malformed query crashes.

### 4. Business Logic Limits ("Max 3 Adoptions Per User")
- The service enforces a hard limit of **3 pets max** per user (`MaxAdoptionsPerUser = 3`).
- When an order is placed (`POST /v2/store/order`) or a pet is registered as adopted:
  - The service checks existing owned pets and active orders for that user.
  - If `existingCount + requestedQuantity > 3`, the request is rejected with `422 Unprocessable Entity`.

---

## Configuration & Environment Variables

| Variable | Default | Description |
|---|---|---|
| `SERVER_ADDR` | `localhost:8080` | Host address and port for the HTTP server |
| `DATABASE_URI` | `mongodb://localhost:27017` | MongoDB connection URI |
| `DATABASE_NAME` | `petstore` | MongoDB database name |
| `ENABLE_CREDENTIALS` | `false` | Set to `true` to authenticate with MongoDB credentials |
| `DATABASE_USERNAME` | *(empty)* | MongoDB username (when `ENABLE_CREDENTIALS=true`) |
| `DATABASE_PASSWORD` | *(empty)* | MongoDB password (when `ENABLE_CREDENTIALS=true`) |
| `GATEWAY_SECRET` | *(empty)* | Optional shared secret; when set, requests must pass matching `X-Gateway-Secret` |
| `ENABLE_AUTH_VALIDATION` | `true` | Enable/disable validation of gateway headers (`X-User-Id`, `X-User-Roles`, `X-User-Email`, `X-Gateway-Secret`) and ABAC enforcement. Default is `true` (enabled). Set to `false` in dev/testing to bypass gateway auth. |

> [!NOTE]
> When `ENABLE_AUTH_VALIDATION=false`, requests without gateway headers (`X-User-Id`, `X-Gateway-Secret`) will not be rejected with `401 Unauthorized`. Instead, a default development user (`dev-user` with `admin` role) is used, allowing direct testing without an API gateway.

---

## KrakenD Integration Guide

Configure KrakenD's `auth/validator` in `krakend.json` to handle JWT validation and claim propagation to this backend:

```json
{
  "version": 3,
  "endpoints": [
    {
      "endpoint": "/v2/pet",
      "method": "POST",
      "extra_config": {
        "auth/validator": {
          "alg": "RS256",
          "jwk_url": "https://www.googleapis.com/oauth2/v3/certs",
          "propagate_claims": [
            ["sub", "X-User-Id"],
            ["roles", "X-User-Roles"],
            ["email", "X-User-Email"]
          ]
        }
      },
      "backend": [
        {
          "url_pattern": "/v2/pet",
          "method": "POST",
          "host": ["http://petstore-backend:8080"],
          "extra_config": {
            "modifier/martian": {
              "header.Modifier": {
                "scope": ["request"],
                "name": "X-Gateway-Secret",
                "value": "your-internal-shared-secret"
              }
            }
          }
        }
      ]
    }
  ]
}
```

---

## Running and Building

### Prerequisites
- Go 1.22+ (configured with Go 1.25)
- MongoDB instance

### Run Locally
```bash
# Start server with default configuration
go run main.go

# Start server with custom MongoDB and Gateway Secret
export SERVER_ADDR="localhost:8080"
export DATABASE_URI="mongodb://localhost:27017"
export DATABASE_NAME="petstore"
export GATEWAY_SECRET="your-internal-shared-secret"
go run main.go
```

### Using the Makefile
A self-documenting [`Makefile`](Makefile) is provided to automate all development tasks:

```bash
make help           # Display all available targets with descriptions
make check          # Run full validation suite (fmt, tidy, lint, test-race, build)
make build          # Build binary into bin/petstore
make run            # Run service locally
make test           # Run unit and security tests
make test-race      # Run tests with race detector
make test-coverage  # Run tests and generate HTML coverage report (bin/coverage.html)
make lint           # Run golangci-lint
make lint-fix       # Automatically apply safe linter fixes
make fmt            # Format Go code
make tidy           # Tidy and verify go.mod / go.sum
make docker-build   # Build optimized ~12 MB multi-stage Docker image
make compose-up     # Launch full stack (MongoDB + Petstore API) via Docker Compose
make compose-down   # Stop and clean up Docker Compose services
make clean          # Remove build artifacts and coverage files
```

### Build with Docker
```bash
# Build the Docker image
make docker-build

# Run the container
make docker-run
```

### Running Tests
Execute the unit and security test suites:
```bash
go test -v -race ./...
```

### Linting & Static Analysis
Run `golangci-lint` locally to verify code quality and security standards:
```bash
# Verify config
golangci-lint config verify

# Run all enabled linters
golangci-lint run ./...
```

---

## CI/CD Pipeline

Continuous Integration is configured via GitHub Actions in [`.github/workflows/ci.yml`](.github/workflows/ci.yml).

The pipeline automatically triggers on `push` and `pull_request` against `main` and `mongo-server` branches, executing three concurrent jobs:
1. **GolangCI-Lint**: Runs `golangci/golangci-lint-action@v6` with `.golangci.yml` covering staticcheck, govet, errcheck, gosec, ineffassign, nilerr, bodyclose, unconvert, and unused.
2. **Unit & Security Tests**: Runs `go test -v -race -cover ./...`.
3. **Build Verification**: Ensures compilation with `go build -v ./...`.
