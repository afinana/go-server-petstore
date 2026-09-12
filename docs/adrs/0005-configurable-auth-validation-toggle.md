# ADR 0005: Configurable Auth Validation Toggle (`ENABLE_AUTH_VALIDATION`)

## Status
Accepted

## Date
2026-09-12

## Context
When microservices mandate specific gateway identity headers (`X-User-Id`, `X-User-Roles`, `X-Gateway-Secret`) and reject unauthenticated requests with `401 Unauthorized`, developer productivity and integration testing can be impacted:
- Local developers testing endpoints via curl, browser, or Postman must manually construct gateway headers on every request.
- Automated testing in isolated test environments (without running a full KrakenD gateway instance) requires mocking gateway headers.
- At the same time, disabling security by default in production introduces critical vulnerabilities.

## Decision
We introduced a centralized configuration variable to control the enforcement of gateway header validation and ABAC:

1. **Configuration Variable**:
   - Environment Variable: `ENABLE_AUTH_VALIDATION` (alias: `ENABLE_HEADER_VALIDATION`).
   - Config struct field: `Config.EnableAuthValidation`.
   - Application method: `app.IsAuthValidationEnabled() bool` and `app.SetAuthValidationEnabled(bool)`.
2. **Secure by Default**:
   - If unset or configured with any value other than `false`, `0`, `off`, or `no`, validation is **strictly enabled**.
3. **Bypass Behavior When Disabled**:
   - Requests without gateway headers are **not** rejected with `401 Unauthorized`.
   - A default permissive development user (`ID: dev-user`, `Roles: [admin]`, `Email: dev@local`) is automatically injected into the request context.
   - Gateway secret verification (`X-Gateway-Secret`) is skipped.
   - Handlers receive an authenticated context with administrative privileges, bypassing ABAC ownership restrictions for local convenience.

## Consequences

### Positive
- **Secure By Default**: Production deployments remain protected by default without requiring manual configuration flags.
- **Superior Developer Ergonomics**: Local development and testing can be conducted by simply setting `ENABLE_AUTH_VALIDATION=false`.
- **Programmatic Control**: Unit and integration test suites can toggle validation state dynamically via `app.SetAuthValidationEnabled(...)`.

### Negative / Trade-offs
- Setting `ENABLE_AUTH_VALIDATION=false` in production would bypass authentication; deployment manifests must ensure this variable is never set to false in public-facing environments.
