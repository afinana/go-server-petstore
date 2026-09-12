# ADR 0003: Strict Input Validation and NoSQL Injection Safeguards

## Status
Accepted

## Date
2026-09-12

## Context
External inputs directly consumed by HTTP handlers presented multiple security and reliability concerns:
1. **Denial of Service (DoS)**: Handlers consumed `r.Body` via `json.NewDecoder` without limiting body size, allowing unbounded payloads to cause memory exhaustion.
2. **Missing Schema & Length Bounds**: Fields like pet name, tags, and category had no length validation, accepting arbitrary payloads or invalid control characters.
3. **Stored XSS / Protocol Manipulation**: `photoUrls` accepted arbitrary strings, including malicious `javascript:` URI schemes.
4. **NoSQL Query Crashes**: Handlers using MongoDB query builders (e.g. `FindByStatus`, `FindByTags`) constructed `$or: []` filters that caused runtime database errors when parameters were empty or invalid.
5. **Path Parameter Tampering**: Parameters like `{petId}` were parsed inconsistently, leading to unhandled errors when malicious strings were passed.

## Decision
We implemented a dedicated input validation module (`petstore/validation.go`) and database query safeguards:

1. **Request Body Size Limiting**:
   - Implemented `LimitRequestBody(w, r)` using `http.MaxBytesReader(w, r.Body, 1<<20)`, capping request bodies at 1 MB.
2. **Field-Level Validation**:
   - **Pet**: Name is required (1–100 chars, no ASCII control characters); status must match allow-list (`available`, `pending`, `sold`); photo URLs must be valid `http` or `https` URLs; tag and category names are bounded to 50 chars.
   - **Order**: `petId` must be a positive integer; quantity must be between 1 and 3; status must match allow-list (`placed`, `approved`, `delivered`, `cancelled`).
3. **Path Parameter Format Validation**:
   - Added `ValidateID(id)` ensuring IDs are either positive 64-bit integers or valid 24-character hexadecimal MongoDB ObjectIDs. Rejects traversal or non-numeric formats with `400 Bad Request`.
4. **NoSQL Query Hardening**:
   - In `PetModel.FindByStatus` and `PetModel.FindByTags`, sanitized slices are queried using the `$in` operator instead of dynamic `$or` arrays, guaranteeing query safety and preventing runtime crashes on empty input.

## Consequences

### Positive
- **Mitigates DoS**: Bounded memory consumption during request parsing.
- **Prevents Injection**: Strongly-typed BSON filters eliminate NoSQL injection risks.
- **Protocol Safety**: Photo URLs are restricted to web protocols (`http`/`https`), eliminating script execution vectors.
- **Predictable Error Responses**: Malformed inputs trigger immediate `400 Bad Request` before invoking database operations.

### Negative / Trade-offs
- Strict validation will reject previously tolerated malformed or oversized payloads, requiring API clients to adhere strictly to the specification.
