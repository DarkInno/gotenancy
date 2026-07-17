$scriptPath = Join-Path $PSScriptRoot 'merge-coverage.ps1'

Describe 'merge-coverage.ps1' {
    It 'sums matching atomic coverage blocks and preserves distinct blocks' {
        $first = Join-Path $TestDrive 'first.out'
        $second = Join-Path $TestDrive 'second.out'
        $output = Join-Path $TestDrive 'combined.out'
        [System.IO.File]::WriteAllLines($first, @(
            'mode: atomic',
            'example.com/project/a.go:1.1,1.2 1 1',
            'example.com/project/b.go:2.1,2.2 2 0'
        ))
        [System.IO.File]::WriteAllLines($second, @(
            'mode: atomic',
            'example.com/project/a.go:1.1,1.2 1 4',
            'example.com/project/c.go:3.1,3.2 1 2'
        ))

        & $scriptPath -Profiles @($first, $second) -Output $output

        ((Get-Content -LiteralPath $output) -join "`n") | Should Be (@(
            'mode: atomic',
            'example.com/project/a.go:1.1,1.2 1 5',
            'example.com/project/b.go:2.1,2.2 2 0',
            'example.com/project/c.go:3.1,3.2 1 2'
        ) -join "`n")
    }
}
