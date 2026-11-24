# Teste de carga do Gatus com k6

O ensaio mede a API `GET /api/v1/endpoints/statuses?page=1&pageSize=50` na revisão `e6a9b061`. O Gatus monitora cinco serviços locais com intervalo de 1 s, mantém até 50 resultados por serviço em memória e usa o cache nativo de 10 s. O k6 consulta o Gatus pela rede interna do Compose.

## Execução

Com Docker Desktop aberto e PowerShell 7.4 ou superior, abra o terminal em `docs/ptoss3/carga` e execute os comandos abaixo.

```powershell
$raizDoClone = git rev-parse --show-toplevel
git -C $raizDoClone archive --format=tar e6a9b061 | docker build --tag ptoss3-carga-gatus:e6a9b061 -
docker compose up -d gatus fixture
.\executar.ps1
docker compose stop gatus fixture
```

O script aguarda os cinco históricos completos e valida as verificações e a gravação dos resumos. Em seguida, roda C1, C2 e C3 com 1, 10 e 20 usuários virtuais, respectivamente, por 2 min cada, e repete a sequência. Cada medição tem uma preparação separada de 20 s com um usuário virtual. Durante as consultas, cada usuário espera 1 s antes de repetir a requisição.

## Evidências

Cada lote fica em `resultados/<data-hora>`, com a conferência dos históricos, a configuração da máquina, os horários e códigos de saída e os seis resumos do k6 em texto e JSON. Os arquivos de validação e preparação registram as checagens anteriores às medições. Os critérios do grupo são p95 abaixo de 500 ms, zero falhas HTTP ou de transporte e 100% das verificações de status e conteúdo aprovadas.
