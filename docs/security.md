# Security

## Tenant Context

- Tenant operations use `core/context.WithTenant`.
- Host operations use `core/context.WithHost`.
- Long-lived jobs should store tenant metadata and rebuild context explicitly.

## GORM Guardrails

- Query, update, delete, row, and count paths add `tenant_id = ?`.
- Create and bulk create fill tenant ID and reject mismatched tenant values.
- `Unscoped` panics in tenant context.
- Raw SQL is rejected in tenant context. `SafeRaw` and `SafeExec` require a context created with `core/context.WithHost`.
- Preload scopes are augmented with tenant filtering.

## Ent And sqlx Guardrails

- Ent integrations expose query filters, mutation filters, and mutation hooks.
- Ent create mutations set `tenant_id` from context and reject mismatched tenant values.
- Ent update and delete mutations receive storage-level tenant predicates.
- sqlx APIs only rewrite simple single-table SELECT/UPDATE/DELETE statements. Complex SQL such as joins, ordering, limits, returning clauses, comments, or multiple statements is rejected with `ErrUnsafeSQL`.

## Active Tenant Enforcement

- HTTP, Gin, Echo, Fiber, Kratos, and gRPC tenant middleware reject non-active tenants by default.
- Active-status guards are also available for trusted contexts created outside middleware.

## Cache Isolation

- Tenant cache keys are prefixed as `t:{tenant_id}:`.
- User-provided keys that already include tenant or global prefixes are rejected.
- Host global keys require explicit opt-in.
- In-memory cache adapters include bounded constructors.

## Error And Log Hygiene

- `web/gin.ErrorHandler` returns generic client errors.
- Web adapters return generic tenant error codes.
- gRPC interceptors return status errors with generic messages.
- `obs.Redact` masks common secret fields before log emission.
- Tenant IDs are emitted as structured observability fields, not embedded into error strings by framework adapters.
