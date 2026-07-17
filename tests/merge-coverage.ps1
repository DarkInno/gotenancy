param(
    [Parameter(Mandatory = $true)]
    [string[]]$Profiles,
    [Parameter(Mandatory = $true)]
    [string]$Output
)

$ErrorActionPreference = 'Stop'
$counts = @{}

foreach ($profile in $Profiles) {
    if (-not (Test-Path -LiteralPath $profile -PathType Leaf)) {
        throw "coverage profile not found: $profile"
    }

    $lines = [System.IO.File]::ReadAllLines((Resolve-Path -LiteralPath $profile))
    if ($lines.Count -eq 0 -or $lines[0] -ne 'mode: atomic') {
        throw "coverage profile must use atomic mode: $profile"
    }

    foreach ($line in $lines | Select-Object -Skip 1) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }

        $parts = $line -split '\s+'
        if ($parts.Count -ne 3) {
            throw "invalid coverage profile entry in ${profile}: $line"
        }

        [long]$count = $parts[2]
        $key = "$($parts[0]) $($parts[1])"
        if ($counts.ContainsKey($key)) {
            $counts[$key] += $count
        } else {
            $counts[$key] = $count
        }
    }
}

$outputDirectory = Split-Path -Parent $Output
if ($outputDirectory -and -not (Test-Path -LiteralPath $outputDirectory -PathType Container)) {
    throw "coverage output directory not found: $outputDirectory"
}

$merged = [System.Collections.Generic.List[string]]::new()
$merged.Add('mode: atomic')
foreach ($key in $counts.Keys | Sort-Object) {
    $merged.Add("$key $($counts[$key])")
}

[System.IO.File]::WriteAllLines($Output, $merged, [System.Text.UTF8Encoding]::new($false))
