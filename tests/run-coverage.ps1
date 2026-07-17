param(
    [Parameter(Mandatory = $true)]
    [string]$Output,
    [string]$IntegrationProfile
)

$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$outputDirectory = Split-Path -Parent $Output
if (-not $outputDirectory) {
    $outputDirectory = $repoRoot
}
if (-not (Test-Path -LiteralPath $outputDirectory -PathType Container)) {
    throw "coverage output directory not found: $outputDirectory"
}
if ($IntegrationProfile -and -not (Test-Path -LiteralPath $IntegrationProfile -PathType Leaf)) {
    throw "integration coverage profile not found: $IntegrationProfile"
}

$profileDirectory = Join-Path $outputDirectory 'saas-module-coverage'
New-Item -ItemType Directory -Path $profileDirectory -Force | Out-Null

function Invoke-ModuleCoverage {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [string]$ModulePath
    )

    $profile = Join-Path $profileDirectory "$Name.out"
    $moduleDirectory = Join-Path $repoRoot $ModulePath
    Push-Location $moduleDirectory
    try {
        & go test -count=1 -covermode=atomic -coverpkg=./... "-coverprofile=$profile" ./... | Out-Host
        if ($LASTEXITCODE -ne 0) {
            throw "go test coverage failed for $ModulePath with exit code $LASTEXITCODE"
        }
    } finally {
        Pop-Location
    }
    return $profile
}

$profiles = @()
$modules = @(
    @{ Name = 'root'; Path = '.' },
    @{ Name = 'data-gorm'; Path = 'data/gorm' },
    @{ Name = 'data-ent'; Path = 'data/ent' },
    @{ Name = 'web-gin'; Path = 'web/gin' },
    @{ Name = 'web-echo'; Path = 'web/echo' },
    @{ Name = 'web-fiber'; Path = 'web/fiber' },
    @{ Name = 'web-kratos'; Path = 'web/kratos' },
    @{ Name = 'rpc-grpc'; Path = 'rpc/grpc' },
    @{ Name = 'cache-redis'; Path = 'cache/redis' },
    @{ Name = 'obs-otel'; Path = 'obs/otel' },
    @{ Name = 'identity-oidc'; Path = 'biz/identity/oidc' },
    @{ Name = 'notification-ses'; Path = 'biz/notification/ses' }
)

foreach ($module in $modules) {
    $profiles += Invoke-ModuleCoverage -Name $module.Name -ModulePath $module.Path
}
if ($IntegrationProfile) {
    $profiles += $IntegrationProfile
}

& (Join-Path $PSScriptRoot 'merge-coverage.ps1') -Profiles $profiles -Output $Output
