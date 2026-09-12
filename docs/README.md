# Project Documentation

Welcome to the documentation for the **Go Petstore Backend Service**.

## Directory Structure

- **[adrs/](adrs/)**: Architecture Decision Records (ADRs) documenting critical architectural and security decisions.
- **[swagger.yaml](../api/swagger.yaml)**: OpenAPI / Swagger 2.0 specification for the Petstore API.

---

## Architecture Summary

The Petstore backend is designed as a microservice operating downstream from an API Gateway (**KrakenD**):

```
Client ──► KrakenD Gateway (Edge TLS, JWT Validation, Rate Limiting)
                 │
                 ├── Headers: X-User-Id, X-User-Roles, X-User-Email, X-Gateway-Secret
                 ▼
           Go Petstore Service (ABAC, Input Validation, Business Limits)
                 │
                 ▼
              MongoDB
```

For setup and deployment instructions, refer to the [Root README](../README.md).
For architectural decisions, see the [Architecture Decision Records](adrs/README.md).
