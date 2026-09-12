# ADR 0007: Developer Tooling, Makefile, and Containerization Improvements

## Status
Accepted

## Date
2026-09-12

## Context
As the Petstore microservice evolves with strict linting, security policies, and gateway architectures, developer workflows need to be streamlined and reproducible:
- Building, formatting, linting, race testing, and coverage analysis were scattered across manual commands.
- The previous `Dockerfile` had inefficient layer caching (`COPY` before `go mod download`), lacked root CA certificates, and ran as root (`uid 0`).
- Local integration testing against MongoDB and KrakenD required manual setup.

## Decision
We introduced standardized developer tooling and containerization improvements:

1. **Self-Documenting Makefile (`Makefile`)**:
   - `make help`: Dynamically lists all targets with color-coded descriptions.
   - `make build` / `make build-linux`: Compiles platform and static Linux binaries into `bin/`.
   - `make test` / `make test-race`: Executes tests with race detection.
   - `make test-coverage`: Generates text and HTML coverage reports (`bin/coverage.html`).
   - `make lint` / `make lint-fix` / `make lint-config`: Enforces `golangci-lint` static analysis.
   - `make fmt` / `make tidy`: Standardizes code formatting and module dependencies.
   - `make check`: Runs complete validation suite (`fmt`, `tidy`, `lint`, `test-race`, `build`).
   - `make docker-build` / `make docker-run`: Manages local container builds.
   - `make compose-up` / `make compose-down`: Automates full-stack local execution.
2. **Multi-Stage Secure Dockerfile (`Dockerfile`)**:
   - Optimizes layer caching by copying `go.mod` and downloading modules prior to copying source code.
   - Embeds root CA certificates (`/etc/ssl/certs/ca-certificates.crt`) for secure outbound TLS.
   - Enforces Principle of Least Privilege by running as non-root user (`USER 65534:65534`).
   - Strips debug symbols (`-ldflags="-w -s"`), resulting in an ultra-compact ~12 MB production image.
3. **Docker Compose & Environment Templates**:
   - `docker-compose.yml`: Spins up MongoDB 7.0 with persistent volume, health checks, and the Petstore backend service.
   - `.env.example`: Provides a documented configuration template.

## Consequences

### Positive
- **Standardized Developer Workflow**: Common commands (`make check`, `make build`, `make test-coverage`) work identically locally and in CI.
- **Fast, Cached Docker Builds**: Source code edits no longer invalidate the dependency download layer.
- **Enhanced Container Security**: Non-root runtime environment with CA certificates.
- **Zero-Friction Local Onboarding**: `make compose-up` boots up a complete working stack in seconds.
