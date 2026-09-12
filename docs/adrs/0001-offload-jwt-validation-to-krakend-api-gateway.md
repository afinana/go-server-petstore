# ADR 0001: Offload JWT Validation to KrakenD API Gateway

## Status
Accepted

## Date
2026-09-12

## Context
In the previous architecture, the Go Petstore service performed in-process JWT validation against Google's OAuth2 JWKS public certificates (`https://www.googleapis.com/oauth2/v3/certs`). This had several drawbacks:
1. **Startup Coupling & Latency**: During application boot (`init()`), the service synchronously fetched remote Google certificates, making service startup dependent on outbound internet connectivity.
2. **Duplicated Perimeter Logic**: Each backend service was responsible for TLS termination, token parsing, signature verification, and expiration checking.
3. **Heavy Cryptographic Dependencies**: Dependencies such as `github.com/MicahParks/keyfunc/v3` and `github.com/golang-jwt/jwt/v5` added binary weight and potential maintenance overhead.
4. **Gateway Deployment**: In modern cloud topologies, an API Gateway such as **KrakenD** is deployed at the edge to handle authentication, rate limiting, and request transformation.

## Decision
We decided to:
1. **Remove In-Process JWT Signature Verification**: Remove `keyfunc` and `jwt/v5` libraries and the Google JWKS fetch logic from the Go service.
2. **Rely on KrakenD for Edge JWT Verification**: KrakenD validates incoming JWT signatures, expiration, and issuer at the perimeter.
3. **Gateway Identity Header Propagation**: KrakenD extracts claims from verified JWTs and passes them downstream via trusted HTTP headers:
   - `X-User-Id` (or `X-Forwarded-User`): Authenticated user / subject identifier (`sub`).
   - `X-User-Roles` (or `X-User-Role`): Comma-delimited list of roles (e.g. `user`, `admin`).
   - `X-User-Email`: Authenticated user email address.
4. **Internal Trust Boundary Defense**: Support an optional `GATEWAY_SECRET` environment variable. When set, requests must contain a matching `X-Gateway-Secret` header, verified using constant-time comparison (`crypto/subtle.ConstantTimeCompare`) to prevent internal attackers from bypassing KrakenD and spoofing headers.

## Consequences

### Positive
- **Faster Startup & Reliability**: No external HTTP calls during application boot; service starts instantly without internet dependency.
- **Lighter Dependencies**: Removed cryptographic JWT parsing libraries from `go.mod`.
- **Centralized Security at Perimeter**: Token validation rules, certificate rotation, and OAuth2 provider configurations are managed centrally in KrakenD.
- **Clear Separation of Concerns**: KrakenD validates identity; the Go backend enforces database-aware authorization and business logic.

### Negative / Trade-offs
- **Internal Network Security Requirement**: The backend relies on trusted identity headers from KrakenD. To prevent header spoofing, either network segregation (private VPC) or the `GATEWAY_SECRET` header check must be enforced in production environments.
