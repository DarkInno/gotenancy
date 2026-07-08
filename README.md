# GoTenancy

ORM-independent Go toolkit for shared-database multi-tenancy with `tenant_id`.

GoTenancy provides tenant context, tenant resolution, data guards, web/RPC middleware, tenant metadata storage, and common SaaS modules.

The default model is simple: every tenant-owned row carries `tenant_id`, and adapters derive the active tenant from `context.Context`.

## Scope

- Shared-database isolation with a required `tenant_id` boundary.
- Host-wide access only through explicit host context.
- GORM, Ent, and sqlx adapters for tenant-aware data access.
- HTTP, Gin, Echo, Fiber, Kratos, and gRPC middleware.
- Tenant lifecycle, plans, subscriptions, quotas, features, RBAC, audit, users, and notifications.

Independent database and hybrid isolation models are not implemented.

## Requirements

- Go `1.23+`.

## Install

```bash
go mod init your-app
go get github.com/DarkInno/gotenancy
```

## Quick Start

```go
db.Use(gormtenant.New(gormtenant.Config{}))

ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
db.WithContext(ctx).Find(&orders)
```

Minimal Ent query filter:

```go
query := client.Order.Query()
_ = enttenant.FilterQuery(ctx, query, enttenant.Config{})
orders, _ := query.All(ctx)
```

For Ent mutations, register `enttenant.Hook(enttenant.Config{})` with the generated client.

See [examples/quickstart](examples/quickstart) for a compiling GORM example.

## Packages

- `core/types`: tenant IDs, tenant metadata, status, and side types.
- `core/context`: tenant and host context, detach, and switch.
- `core/resolver`: header, cookie, query, domain, token-claim, and composite resolvers.
- `core/store`: memory store, paginated list filters, memory cache, cached store decorator, and `database/sql` store.
- `data`: ORM-independent tenant filter condition.
- `data/gorm`: GORM plugin, guard suite, host-only `SafeRaw`/`SafeExec`, `BulkCreate`, and delete APIs.
- `data/ent`: Ent selector predicate, query filter, mutation filter, and hook APIs.
- `data/sqlx`: tenant-filtered APIs for simple single-table SELECT/UPDATE/DELETE statements.
- `saas/tenant`: tenant lifecycle state machine.
- `saas/plan`: plan CRUD.
- `saas/subscription`: subscription lifecycle and billing hook.
- `saas/quota`: quota checking and atomic consumption.
- `saas/feature`: plan defaults plus tenant-level feature overrides.
- `web/*`: tenant middleware and guards for net/http, Gin, Echo, Fiber, and Kratos.
- `rpc/grpc`: gRPC unary and stream tenant interceptors.
- `migration`: tenant column and index planning.
- `cache`: tenant-scoped cache wrapper and memory adapters.
- `obs`: observability fields and redaction.
- `biz/*`: user, RBAC, audit, and notification modules.

## Verification

```bash
go test ./...
go vet ./...
go test -race ./...
```

On Windows, `go test -race` requires cgo and a C compiler. Without local cgo, run race tests in Docker:

```bash
docker run --rm -v "${PWD}:/workspace" -w /workspace -e CGO_ENABLED=1 -e GOFLAGS=-mod=readonly golang:1.23 go test -race ./...
```

Optional database integration tests:

```bash
GOTENANCY_MYSQL_DSN='<mysql-dsn>' go test ./core/store -run TestSQLStoreMySQLIntegration -count=1
GOTENANCY_POSTGRES_DSN='<postgres-dsn>' go test ./core/store -run TestSQLStorePostgresIntegration -count=1
GOTENANCY_MYSQL_DSN='<mysql-dsn>' go test ./data/gorm -run TestMySQLIntegrationEnforcesTenantIsolation -count=1
```

## Project Layout

```text
core/          Tenant context, resolver, store, and types
data/          Data filtering contracts and adapters
saas/          Tenant lifecycle, plan, subscription, quota, and feature modules
web/           Web framework and net/http integration
migration/     Tenant schema migration planning
cache/         Tenant-aware cache abstractions
rpc/           RPC metadata propagation
obs/           Observability fields and redaction
biz/           User, RBAC, audit, and notification modules
examples/      Runnable examples
tests/         Security, cache, and concurrency tests
docs/          API, security, and compatibility notes
```

## Compatibility

See [docs/compatibility.md](docs/compatibility.md).

## License

[Apache License 2.0](LICENSE)
