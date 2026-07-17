param([switch]$KeepServices)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$composeFile = Join-Path $PSScriptRoot 'compose.yaml'

function Invoke-Checked {
    param([string]$File, [string[]]$Arguments)
    & $File @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$File $($Arguments -join ' ') failed with exit code $LASTEXITCODE"
    }
}

Push-Location $repoRoot
try {
    Invoke-Checked docker @('compose', '-f', $composeFile, 'up', '-d', '--wait')
    $env:GOTENANCY_MYSQL_DSN = 'root:gotenancy@tcp(127.0.0.1:33067)/gotenancy_test?parseTime=true&timeout=3s&readTimeout=3s&writeTimeout=3s'
    $env:GOTENANCY_POSTGRES_DSN = 'postgres://gotenancy:gotenancy@127.0.0.1:55432/gotenancy_test?sslmode=disable'
    $env:GOTENANCY_REDIS_ADDR = '127.0.0.1:56379'
    $env:GOTENANCY_REDIS_DB = '15'

    Invoke-Checked go @('test', './data/gorm', '-run', '^TestMySQLIntegrationEnforcesTenantIsolation$', '-count=1')
    Push-Location (Join-Path $repoRoot 'tests/db')
    try {
        Invoke-Checked go @('test', './...', '-run', '^TestSQLStore(MySQL|Postgres)Integration$', '-count=1')
    } finally {
        Pop-Location
    }
    Invoke-Checked go @('test', './cache', '-run', '^TestRedisCacheIntegration$', '-count=1')
} catch {
    & docker compose -f $composeFile logs --no-color
    throw
} finally {
    Remove-Item Env:GOTENANCY_MYSQL_DSN, Env:GOTENANCY_POSTGRES_DSN, Env:GOTENANCY_REDIS_ADDR, Env:GOTENANCY_REDIS_DB -ErrorAction SilentlyContinue
    if (-not $KeepServices) {
        & docker compose -f $composeFile down --volumes --remove-orphans
    }
    Pop-Location
}
