# Ambiente exploratório da PTOSS-3

Na branch `ptoss3` do fork, abra o terminal na raiz do clone e execute com o Docker Desktop aberto:

```powershell
Set-Location -LiteralPath '.\docs\ptoss3\ambiente'
docker compose up -d --build
```

Abra `http://127.0.0.1:8080`. O painel apresenta seis serviços, incluindo `painel-falha`, com falha intencional. Se a porta 8080 estiver ocupada, defina `$env:PTOSS3_PORT = '18080'` antes de subir e acesse `http://127.0.0.1:18080`.

Siga o plano e o modelo recebidos da equipe. Use a pasta `docs/ptoss3/registros/EX0X` do seu identificador para o registro e as evidências. Ao concluir, restaure os arquivos de configuração e de apoio usados na sessão e execute `docker compose stop` na pasta `ambiente`.

## Registros e evidências

| Sessão | Funcionalidade | Registro |
| --- | --- | --- |
| EX01 | Falha e recuperação no monitoramento HTTP | [Registro e evidências](registros/EX01) |
| EX02 | Busca e filtro de falhas no painel | [Registro e evidências](registros/EX02) |
| EX03 | Validação de campo numérico no JSON | [Registro e evidências](registros/EX03) |
| EX04 | Recarga da configuração durante a execução | [Registro e evidências](registros/EX04) |
| EX05 | Paginação do histórico de um serviço | [Registro e evidências](registros/EX05) |

O [ensaio com k6](carga/README.md) tem configuração, scripts e [resumos das seis medições](carga/resultados/20261123-120850).
