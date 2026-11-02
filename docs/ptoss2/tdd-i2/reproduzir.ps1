param(
    [string]$GoExecutable = 'go',
    [string]$OutputDirectory
)
$ErrorActionPreference = 'Stop'
$goCommand = (Get-Command $GoExecutable -ErrorAction Stop).Source
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../../..')).Path
$source = Join-Path $repo 'config/endpoint/ui'
$output = if ($OutputDirectory) { [System.IO.Path]::GetFullPath($OutputDirectory) } else { Join-Path $PSScriptRoot ('evidencias/execucao-' + (Get-Date -Format 'yyyyMMdd-HHmmss')) }
New-Item -ItemType Directory -Path $output | Out-Null
$temporaryRoot = Join-Path ([System.IO.Path]::GetTempPath()) ('ptoss2-i2-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temporaryRoot | Out-Null
$stages = @(
    @{ Name = '00-antes'; State = '00-original.go.txt'; Tests = @('ui_test.go'); Run = '.'; Exit = 0 },
    @{ Name = '01-caixa-preta'; State = '00-original.go.txt'; Tests = @('ui_test.go', 'ui_badge_ptoss2_blackbox_test.go'); Run = '.'; Exit = 0 },
    @{ Name = '02-caixa-preta-branca'; State = '00-original.go.txt'; Tests = @('ui_test.go', 'ui_badge_ptoss2_blackbox_test.go', 'ui_badge_ptoss2_whitebox_test.go'); Run = '.'; Exit = 0 },
    @{ Name = '03-tdd1-red'; State = '00-original.go.txt'; Tests = @('ui_badge_ptoss2_blackbox_test.go', 'ui_badge_ptoss2_tdd1_test.go'); Run = '^TestBadgePTOSS2TDD1$'; Exit = 1 },
    @{ Name = '04-tdd1-green'; State = '01-tdd1.go.txt'; Tests = @('ui_test.go', 'ui_badge_ptoss2_blackbox_test.go', 'ui_badge_ptoss2_whitebox_test.go', 'ui_badge_ptoss2_tdd1_test.go'); Run = '.'; Exit = 0 },
    @{ Name = '05-tdd2-red'; State = '01-tdd1.go.txt'; Tests = @('ui_badge_ptoss2_blackbox_test.go', 'ui_badge_ptoss2_tdd2_test.go'); Run = '^TestBadgePTOSS2TDD2$'; Exit = 1 },
    @{ Name = '06-tdd2-green'; State = $null; Tests = @('ui_test.go', 'ui_badge_ptoss2_blackbox_test.go', 'ui_badge_ptoss2_whitebox_test.go', 'ui_badge_ptoss2_tdd1_test.go', 'ui_badge_ptoss2_tdd2_test.go'); Run = '.'; Exit = 0 }
)
$log = Join-Path $output 'execucoes.txt'
@('Reproducao posterior; nao e evidencia de execucao durante o desenvolvimento.', ('Data: ' + (Get-Date -Format o)), ('Revisao: ' + (& git -C $repo rev-parse HEAD)), (& $goCommand version)) | Set-Content $log
foreach ($stage in $stages) {
    $stageDir = Join-Path $temporaryRoot $stage.Name
    New-Item -ItemType Directory -Path $stageDir | Out-Null
    $product = if ($stage.State) { Join-Path $PSScriptRoot ('estados/' + $stage.State) } else { Join-Path $source 'ui.go' }
    Copy-Item -LiteralPath $product -Destination (Join-Path $stageDir 'ui.go')
    foreach ($test in $stage.Tests) { Copy-Item -LiteralPath (Join-Path $source $test) -Destination $stageDir }
    "module ptoss2badge`n`ngo 1.26.3`n" | Set-Content (Join-Path $stageDir 'go.mod')
    Get-ChildItem -LiteralPath $stageDir -Filter '*.go' | ForEach-Object {
        $hash = Get-FileHash -LiteralPath $_.FullName -Algorithm SHA256
        Add-Content $log ($stage.Name + '/' + $_.Name + ' SHA256 ' + $hash.Hash)
    }
    Push-Location $stageDir
    try {
        $commandText = 'go test -count=1 -v -run ' + $stage.Run + ' -coverprofile=coverage.out .'
        Add-Content $log ($stage.Name + ': ' + $commandText)
        $previousPreference = $ErrorActionPreference
        $ErrorActionPreference = 'Continue'
        $testOutput = & $goCommand test -count=1 -v -run $stage.Run '-coverprofile=coverage.out' . 2>&1
        $testExit = $LASTEXITCODE
        $ErrorActionPreference = $previousPreference
        $testOutput | Out-Host
        $testOutput | Out-String | Add-Content $log
        Add-Content $log ('Exit: ' + $testExit + '; esperado: ' + $stage.Exit)
        if ($testExit -ne $stage.Exit) { throw ('Resultado inesperado em ' + $stage.Name + '; consulte execucoes.txt') }
        if ($stage.Exit -eq 1) {
            $joined = $testOutput | Out-String
            $expectedFailure = if ($stage.Name -eq '03-tdd1-red') { 'validation panicked instead of returning configuration error' } else { 'error = <nil>; want invalid response time badge configuration' }
            if (!$joined.Contains($expectedFailure)) { throw ('Red sem a falha esperada: ' + $stage.Name) }
        } else {
            & $goCommand tool cover '-func=coverage.out' | Tee-Object -FilePath (Join-Path $output ($stage.Name + '.cover-func.txt')) | Out-Host
            if ($LASTEXITCODE -ne 0) { throw 'Falha ao obter cobertura' }
        }
    } finally { Pop-Location }
}
Write-Output ('Resultados: ' + $output)
