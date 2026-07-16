# 兼容性

[EN](compatibility.md) | [中文](compatibility.zh-CN.md)

## Go

- 模块语言版本：Go `1.24`。
- `go.mod` 将其记录为 `go 1.24.0`。
- CI 测试任务应覆盖 Go `1.24.x` 和 Go `1.26.x`；lint 和漏洞扫描在已修复的 Go `1.26.5+` 工具链上运行。

由于 OIDC 路径所需的已修复 `github.com/go-jose/go-jose/v4` 版本要求 Go `1.24.0`，因此不再支持 Go `1.23`。

除非作出明确的兼容性决策，模块不应引入要求 Go `1.25+` 的依赖。

## 隔离模型

GoTenancy 支持以必需的 `tenant_id` 边界实现共享数据库隔离。

独立数据库和混合隔离模型不属于当前 API 的范围。

## 适配器

| 适配器 | 依赖 |
|---|---|
| GORM v2 | `gorm.io/gorm` v1.31.2 |
| Ent | `entgo.io/ent` v0.14.1 |
| Gin | `github.com/gin-gonic/gin` v1.9.1 |
| Echo | `github.com/labstack/echo/v4` v4.13.4 |
| Fiber | `github.com/gofiber/fiber/v2` v2.52.13 |
| Kratos | `github.com/go-kratos/kratos/v2` v2.9.2 |
| gRPC | `google.golang.org/grpc` v1.75.1 |
| OIDC | `github.com/coreos/go-oidc/v3` v3.15.0 和 `golang.org/x/oauth2` v0.30.0 |
| Redis 缓存 | `github.com/redis/go-redis/v9` v9.21.0 |

`core/` 保持不导入 GORM、Ent、sqlx、Redis 和 web 框架。

## SQLStore

`core/store.SQLStore` 支持：

- MySQL/SQLite 占位符：`?`
- PostgreSQL 占位符：`$1`、`$2`、...

对于 PostgreSQL，请使用 `WithSQLDialect(SQLDialectPostgres)`。

可选集成测试使用：

- `GOTENANCY_MYSQL_DSN`
- `GOTENANCY_POSTGRES_DSN`

SQLStore 数据库集成测试位于 `tests/db`，不属于默认 CI 门禁。

只有设置 `GOTENANCY_REDIS_ADDR` 时才运行可选的 Redis 缓存集成测试。`GOTENANCY_REDIS_PASSWORD` 和 `GOTENANCY_REDIS_DB` 为可选项。

## 验证

```bash
go test ./...
go vet ./...
go test -race ./...
go list -m -f '{{.Path}} {{.GoVersion}}' all
go run golang.org/x/vuln/cmd/govulncheck@v1.5.0 ./...
```

在 Windows 上如果本地没有 cgo，请在 Docker 中运行 race 测试：

```bash
docker run --rm -v "${PWD}:/workspace" -w /workspace -e CGO_ENABLED=1 -e GOFLAGS=-mod=readonly golang:1.24 go test -race ./...
```
