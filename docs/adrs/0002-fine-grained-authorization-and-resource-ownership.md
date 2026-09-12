# ADR 0002: Fine-Grained Authorization (ABAC) and Resource Ownership

## Status
Accepted

## Date
2026-09-12

## Context
While the API Gateway (KrakenD) can verify token authenticity and coarse-grained role checks (RBAC), it is **stateless** and does not possess access to the application database. Consequently, KrakenD cannot verify resource-level ownership:
- *Example*: Is User A authorized to edit or delete Pet B?
- *Example*: Is User A authorized to inspect or cancel Order C placed by User B?

In the previous implementation:
- `Pet` had no owner reference; any authenticated caller could delete or update any pet.
- `Order` had no user association; orders were stored in a global slice without user context.
- This exposed the system to **Insecure Direct Object Reference (IDOR)** vulnerabilities (OWASP Top 10 / CWE-639).

## Decision
We decided to enforce **Attribute-Based Access Control (ABAC)** and resource-level ownership directly within the Go backend service:

1. **Model Updates**:
   - Added `ownerId` to the `Pet` model (`primitive.ObjectID` or string ID).
   - Added `userId` to the `Order` model.
2. **Context Association**:
   - On resource creation (`POST /v2/pet`), `pet.OwnerID` is assigned from the caller's verified `X-User-Id`.
   - On order creation (`POST /v2/store/order`), `order.UserId` is assigned from `X-User-Id`.
3. **Authorization Policy**:
   - Handlers verify ownership prior to data mutation or sensitive queries:
     - **Owners**: A user can modify or delete resources where `resource.OwnerID == user.ID`.
     - **Administrators**: Users with role `admin` (`X-User-Roles: admin`) can manage all resources regardless of ownership.
     - **Unowned / Legacy Resources**: Resources created without an owner remain manageable to ensure backward compatibility.
     - **Unauthorized Access**: Returns HTTP `403 Forbidden`.
4. **Order Inventory Isolation**:
   - `GET /v2/store/inventory` returns all orders for administrators, but strictly filters to the authenticated user's orders for standard users.

## Consequences

### Positive
- **IDOR Mitigation**: Eliminates unauthorized cross-user modifications and data tampering.
- **Principle of Least Privilege**: Users can only access and modify their own data.
- **Decoupled Architecture**: Gateway remains fast and stateless while the backend leverages database state for fine-grained authorization.

### Negative / Trade-offs
- **Database Lookups Before Mutation**: Operations such as `UpdatePet` and `DeletePet` require querying the existing resource from MongoDB to verify ownership before performing mutations.
