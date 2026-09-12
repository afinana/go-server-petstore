# ADR 0004: Enforce Business Logic Limits for Pet Adoption

## Status
Accepted

## Date
2026-09-12

## Context
In business applications, domain constraints must be enforced at the business logic layer to prevent abuse, resource hoarding, and state corruption:
- *Requirement*: A user can only adopt a maximum of **3 pets**.
- In the existing implementation, order creation was unconstrained: users could place an unlimited number of orders with arbitrary quantities (including negative numbers or millions of pets).
- An API Gateway cannot enforce this rule because the calculation requires checking the current number of pets adopted/owned by the user in the database.

## Decision
We implemented domain-level adoption limits within the Petstore application logic:

1. **Adoption Constant**:
   - Defined `MaxAdoptionsPerUser = 3` in `petstore/validation.go`.
2. **Order Quantity Constraints**:
   - In `ValidateOrder`, each order's `quantity` must be between 1 and 3.
3. **Aggregate Ownership Evaluation**:
   - In `PlaceOrder` (`POST /v2/store/order`):
     - The service aggregates the user's active orders (`order.Status != "cancelled"`) and pets directly registered to the user in MongoDB (`CountByOwner`).
     - If `currentAdopted + requestedQuantity > MaxAdoptionsPerUser`, the request is rejected with HTTP `422 Unprocessable Entity`.
4. **Pet Registration Enforcement**:
   - In `AddPet` (`POST /v2/pet`), when non-admin users register a pet, their current pet count is verified against `MaxAdoptionsPerUser`.
5. **Administrative Exemption**:
   - Users with role `admin` are exempt from the adoption cap to facilitate administrative store management and inventory provisioning.

## Consequences

### Positive
- **Guaranteed Business Rule Adherence**: Impossible for standard users to exceed the maximum quota of 3 adopted pets.
- **Clear Semantic Error Responses**: Returns HTTP `422 Unprocessable Entity` with explanatory messaging detailing the current count and maximum limit.
- **Administrative Flexibility**: Admins retain the ability to manage inventory without quota blocks.

### Negative / Trade-offs
- Additional database aggregate count queries on order creation.
