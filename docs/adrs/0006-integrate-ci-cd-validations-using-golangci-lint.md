# ADR 0006: Integrate CI/CD Validations Using golangci-lint

## Status
Accepted

## Date
2026-09-12

## Context
Maintaining high code quality, security posture, and consistency across a Go microservice codebase requires automated continuous integration (CI) checks:
- Preventing unhandled errors (`errcheck`).
- Catching security vulnerabilities like integer overflows or log injection (`gosec`).
- Ensuring correctness, avoiding dead code, and preventing nil-error bugs (`staticcheck`, `unused`, `nilerr`, `govet`).
- Enforcing that tests and builds pass automatically on every pull request and push.

## Decision
We integrated **golangci-lint** into the repository and configured continuous integration via GitHub Actions:

1. **Linter Configuration (`.golangci.yml`)**:
   - Utilizes `version: "2"` schema.
   - Enables core quality and security linters: `bodyclose`, `errcheck`, `gosec`, `govet`, `ineffassign`, `misspell`, `nilerr`, `staticcheck`, `unconvert`, and `unused`.
2. **GitHub Actions Workflow (`.github/workflows/ci.yml`)**:
   - Executes automatically on `push` and `pull_request` to `main` and `mongo-server` branches.
   - **`lint` job**: Uses official `golangci/golangci-lint-action@v6` to run linters with automated caching and parallel execution.
   - **`test` job**: Runs `go test -v -race -cover ./...` to verify unit and security regression tests with race detector enabled.
   - **`build` job**: Runs `go build -v ./...` to ensure project compilation across packages.

## Consequences

### Positive
- **Automated Regression Prevention**: Regressions, security flaws (e.g. CWE log injection, integer overflows), and unhandled errors are automatically flagged prior to merging.
- **Fast Developer Feedback**: Developers can run `golangci-lint run ./...` locally before pushing code.
- **Seamless GitHub Integration**: Annotations appear directly on pull requests when issues are detected.

### Negative / Trade-offs
- CI workflow execution time increases slightly due to running linters and race-detector tests.
