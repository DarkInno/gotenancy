module github.com/DarkInno/gotenancy/tests/db

go 1.24.0

replace github.com/DarkInno/gotenancy => ../..

require (
	github.com/DarkInno/gotenancy v0.0.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/lib/pq v1.10.9
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/coreos/go-oidc/v3 v3.15.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
)
