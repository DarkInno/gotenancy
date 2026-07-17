# 架构

[EN](architecture.md) | [中文](architecture.zh-CN.md)

SaaS 是一个组装到宿主 Go 应用中的库；它本身不运行 HTTP/gRPC 服务、消息队列 Broker、Broker 客户端连接或物理部署基础设施。本图展示了该模块实现的集成边界以及常规的租户作用域请求路径。存储、消息队列和外部系统节点由宿主选择和配置；它们是受支持的集成点，而不是本仓库部署的服务。

```mermaid
flowchart TB
    caller["HTTP/gRPC clients<br/>消息生产者、CLI 和后台任务"]

    subgraph host["Host Go application"]
        web["HTTP integration<br/>web/http, Gin, Echo, Fiber, Kratos"]
        grpc["gRPC integration<br/>rpc/grpc interceptors"]
        mq["消息队列集成<br/>rpc/mq 载体；客户端循环由宿主负责"]
        direct["Direct invocation<br/>workers, CLI, application services"]
        app["Handlers and application services<br/>propagate context.Context"]
        modules["SaaS and business modules<br/>tenant、plan、subscription、quota、feature、onboarding 和 biz/*"]
        telemetry["Observability helpers<br/>obs; host configures exporters"]
    end

    subgraph boundary["Tenant boundary"]
        resolver["HTTP tenant resolver<br/>header, cookie, query, domain, token"]
        carrier["RPC metadata carrier<br/>rpc"]
        mqCarrier["MQ 消息 header 载体<br/>rpc/mq：NATS、RabbitMQ、Kafka"]
        tenantStore["Tenant metadata store<br/>core/store.Store"]
        tenantContext["Tenant or explicit host context<br/>core/context"]
        tenantTypes["Tenant IDs, metadata, and state<br/>core/types"]
    end

    subgraph isolation["Tenant isolation and cache"]
        dataFilter["Context-derived data filter<br/>data"]
        gorm["GORM callbacks and guards<br/>data/gorm"]
        ent["Ent predicates, filters, and hooks<br/>data/ent"]
        sqlx["Safe simple-statement filtering<br/>data/sqlx"]
        cache["Tenant-scoped cache keys<br/>cache.TenantCache"]
        migration["Tenant column and index DDL planner<br/>migration"]
    end

    subgraph adapters["Host-selected storage and external adapters"]
        stores["Memory or host-provided database/sql stores<br/>core、生命周期模块和 biz"]
        database["由宿主维护的共享应用数据库和 Schema<br/>tenant-owned rows include tenant_id"]
        redis["Host-provided optional Redis cache"]
        broker["由宿主管理的 NATS、RabbitMQ 或 Kafka"]
        idp["OIDC identity provider"]
        delivery["SMTP, SES, Resend, or webhook"]
    end

    caller --> web
    caller --> grpc
    caller --> mq
    caller --> direct
    web --> resolver
    grpc --> carrier
    mq --> mqCarrier
    resolver --> tenantStore
    carrier --> tenantStore
    mqCarrier --> tenantStore
    direct --> tenantContext
    tenantStore --> tenantContext
    tenantTypes -. "defines" .-> resolver
    tenantTypes -. "defines" .-> tenantStore
    tenantTypes -. "defines" .-> tenantContext
    tenantContext --> app
    app --> modules
    app --> telemetry
    tenantContext --> dataFilter
    app -. "selected adapter" .-> gorm
    app -. "selected adapter" .-> ent
    app -. "selected adapter" .-> sqlx
    dataFilter -. "used by" .-> gorm
    dataFilter -. "used by" .-> ent
    dataFilter -. "used by" .-> sqlx
    app -. "optional" .-> cache
    tenantContext -. "scopes" .-> cache
    modules --> stores
    tenantStore --> stores
    stores --> database
    gorm --> database
    ent --> database
    sqlx --> database
    cache --> redis
    app --> mq
    mq --> broker
    migration -. "plans; does not execute" .-> database
    modules -. "optional OIDC bridge" .-> idp
    modules -. "optional notifications" .-> delivery
```

## 租户作用域请求路径

```mermaid
flowchart LR
    inbound["传入的 HTTP 请求、gRPC 元数据<br/>或 MQ 消息 header"]
    resolve["Resolve tenant ID"]
    lookup{"Tenant exists<br/>and is active?"}
    reject["Reject with tenant_required,<br/>tenant_forbidden, or tenant_inactive"]
    scoped["Attach tenant to context.Context"]
    handler["Host handler or service"]
    guard["Data adapter or cache wrapper"]
    protected["tenant_id predicate or<br/>tenant-prefixed cache key"]

    inbound --> resolve --> lookup
    lookup -->|no| reject
    lookup -->|yes| scoped --> handler --> guard --> protected
```

## 边界规则

- HTTP 和 gRPC 集成会解析租户、加载其元数据，并要求租户处于活跃状态，之后才将控制权交给宿主应用。
- 配置可选部署解析器后，这些集成会在租户查询成功后解析其逻辑部署单元，并将其写入同一个 `context.Context`。该目录不会选择数据库连接、路由流量或搬迁数据；这些操作仍由宿主负责。参阅[部署单元](deployment.zh-CN.md)。
- `rpc/mq` 只适配 NATS、RabbitMQ 和 Kafka 的消息 header。出站工作中，宿主从已经建立的租户上下文调用 `rpc.InjectTenant`；入站工作中，宿主调用 `rpc.ExtractTenant`，通过 `core/store.Store` 加载租户、确认其处于活跃状态，然后在分发消息前调用 `core/context.WithTenant`。
- 所有 Broker I/O 和消息策略均由宿主负责。MQ 适配器不会建立连接、发布或消费消息、确认消息、执行重试或死信处理，也不会校验租户元数据。
- `context.Context` 是作用域载体。后台任务必须显式建立租户上下文；全局主机操作必须使用有意为之的 `core/context.WithHost` 路径。
- GORM、Ent 和 sqlx 适配器从该上下文派生数据边界。在受支持的共享数据库、共享 Schema 模型中，租户所有的行都带有 `tenant_id`；本模块不实现按租户独立数据库、独立 Schema 或混合隔离。
- 存储可以使用内存实现，也可以使用宿主提供的 SQL 连接。Redis 是可选的、由宿主提供的缓存适配器，而不是租户隔离的来源。
- `migration.Planner` 生成租户感知的 DDL 和 seed 语句；它从不执行迁移。

有关包级接口，请参阅 [API 参考](api.zh-CN.md)；有关详细的防护行为，请参阅[安全性](security.zh-CN.md)。
