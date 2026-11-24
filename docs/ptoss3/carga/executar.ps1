$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
Set-Location -LiteralPath $PSScriptRoot

$lote = Get-Date -Format 'yyyyMMdd-HHmmss'
$pasta = Join-Path 'resultados' $lote
New-Item -ItemType Directory -Path $pasta | Out-Null

$validarHistoricos = @'
from datetime import datetime, timezone
from json import dumps, load
from time import monotonic, sleep
from urllib.request import urlopen

rota = "http://gatus:8080/api/v1/endpoints/statuses?page=1&pageSize=50"
esperados = {f"servico-{numero}" for numero in range(1, 6)}
with urlopen("http://fixture:8000/ok", timeout=4) as apoio:
    assert apoio.status == 200 and apoio.read() == b"OK", "Resposta inesperada do apoio"

limite = monotonic() + 100
while monotonic() < limite:
    try:
        with urlopen(rota, timeout=4) as resposta:
            dados = load(resposta)
        historicos = {servico["name"]: len(servico["results"]) for servico in dados}
        if (
            len(dados) == 5 and set(historicos) == esperados
            and all(servico["group"] == "seminario" for servico in dados)
            and all(total == 50 for total in historicos.values())
            and all(servico["results"][-1]["success"] for servico in dados)
        ):
            print(dumps({"emUtc": datetime.now(timezone.utc).isoformat(),
                         "rota": rota, "http": 200, "grupo": "seminario",
                         "historicos": historicos, "ultimosChecksSaudaveis": True}, indent=2))
            break
    except (OSError, ValueError):
        pass
    sleep(2)
else:
    raise SystemExit("A API não apresentou os cinco históricos completos no prazo.")
'@

Write-Host "Lote $lote. Conferindo os cinco serviços e seus históricos."
$prontidao = docker compose exec -T fixture python -c $validarHistoricos
if ($LASTEXITCODE -ne 0) { throw 'Falha na conferência dos dados de entrada.' }
$prontidao | Set-Content -LiteralPath (Join-Path $pasta 'historicos.json') -Encoding utf8

$versaoK6 = docker compose run -T --rm --no-deps k6 version
if ($LASTEXITCODE -ne 0) { throw 'Falha ao confirmar a versão do k6.' }
$ambiente = [ordered]@{
    lote = $lote
    preparadoEm = (Get-Date).ToString('o')
    revisaoGatus = (git rev-parse --verify 'e6a9b061^{commit}')
    revisaoDaBranch = (git rev-parse HEAD)
    cpu = (Get-CimInstance Win32_Processor | Select-Object Name, NumberOfCores, NumberOfLogicalProcessors)
    ramBytes = (Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
    sistema = (Get-CimInstance Win32_OperatingSystem | Select-Object Caption, Version, BuildNumber)
    docker = (docker version --format '{{.Server.Version}}')
    compose = (docker compose version --short)
    motor = (docker info --format '{{.OSType}}/{{.Architecture}}, {{.NCPU}} vCPUs, {{.MemTotal}} bytes de RAM')
    k6 = $versaoK6
    imagemK6 = (docker image inspect grafana/k6:1.5.0 --format '{{index .RepoDigests 0}}')
    imagemGatus = (docker image inspect ptoss3-carga-gatus:e6a9b061 --format '{{.Id}}')
    python = (docker compose exec -T fixture python --version)
    imagemPython = (docker image inspect 'python:3.12-alpine@sha256:c5ba3deef00c7758561a0f03c26f44f10589263d40e732b8748e0d1f38f6dd89' --format '{{index .RepoDigests 0}}')
    rede = (docker network inspect ptoss3-carga_default --format '{{.Driver}}, internal={{.Internal}}')
    limites = 'Compose sem limites individuais para os três serviços.'
    outrosContainersAtivos = @(docker ps --format '{{.Label "com.docker.compose.project"}}' | Where-Object { $_ -ne 'ptoss3-carga' }).Count
}
$ambiente | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $pasta 'ambiente.json') -Encoding utf8

Write-Host 'Validando o script e a exportação antes das medições.'
docker compose run -T --rm --no-deps k6 run --quiet --summary-export "/resultados/$lote/validacao-k6.json" -e PREPARACAO=false -e VUS=1 -e DURACAO=3s /scripts/consulta-estados.js 2>&1 |
    Set-Content -LiteralPath (Join-Path $pasta 'validacao-k6.txt') -Encoding utf8
if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath (Join-Path $pasta 'validacao-k6.json'))) {
    throw "A validação do k6 falhou. Consulte $pasta."
}

$cenarios = [ordered]@{ C1 = 1; C2 = 10; C3 = 20 }
$preparacoes = Join-Path $pasta 'preparacoes.txt'
$execucoes = Join-Path $pasta 'execucoes.jsonl'
foreach ($repeticao in 1, 2) {
    foreach ($cenario in $cenarios.Keys) {
        $nome = "$cenario-R$repeticao"
        $vus = $cenarios[$cenario]
        Write-Host "$nome. Preparação de 20 s."
        Add-Content -LiteralPath $preparacoes -Value "--- $nome ---" -Encoding utf8
        docker compose run -T --rm --no-deps k6 run --quiet -e PREPARACAO=true -e VUS=1 -e DURACAO=20s /scripts/consulta-estados.js 2>&1 |
            Add-Content -LiteralPath $preparacoes -Encoding utf8
        if ($LASTEXITCODE -ne 0) { throw "Falha na preparação de $nome." }

        Write-Host "$nome. Medição de 2 min com $vus VU(s)."
        $inicio = (Get-Date).ToString('o')
        docker compose run -T --rm --no-deps k6 run --quiet --summary-export "/resultados/$lote/$nome.json" -e PREPARACAO=false -e "VUS=$vus" -e DURACAO=2m /scripts/consulta-estados.js 2>&1 |
            Set-Content -LiteralPath (Join-Path $pasta "$nome.txt") -Encoding utf8
        $codigo = $LASTEXITCODE
        [ordered]@{ medicao = $nome; vus = $vus; duracao = '2m'; inicio = $inicio; fim = (Get-Date).ToString('o'); codigoK6 = $codigo } |
            ConvertTo-Json -Compress | Add-Content -LiteralPath $execucoes -Encoding utf8
        if ($codigo -notin 0, 99 -or -not (Test-Path -LiteralPath (Join-Path $pasta "$nome.json"))) {
            throw "Erro de execução em $nome. As evidências foram preservadas em $pasta."
        }
        Write-Host "$nome concluído. Código k6 $codigo."
    }
}

Write-Host "Medições registradas em $pasta."
