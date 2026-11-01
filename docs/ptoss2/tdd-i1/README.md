# PTOSS-2 — I1: TDD do resolvedor DNS

**Responsável:** surpesaiajin · **Acompanhamento:** [issue #2 deste fork](https://github.com/surpesaiajin/gatus/issues/2).

## Comportamentos desenvolvidos

O [parser](../../../client/config.go#L155) é usado por `Config.ValidateAndSetDefaults` para validar `DNSResolver` antes do uso pelo cliente. Os dois ciclos evoluem comportamentos distintos:

- **V01:** rejeitar sufixos após a porta, como `tcp://localhost:53abc`, preservando formatos válidos e portas de 1 a 65535.
- **V02:** aceitar somente `tcp`, `tcp4`, `tcp6`, `udp`, `udp4` e `udp6`; rejeitar redes incompatíveis, como `http`. Essa restrição foi adotada pela equipe como melhoria de validação.

Os [testes](../../../client/config_dns_resolver_ptoss2_tdd_test.go) verificam o parser, a validação pública, os componentes retornados e a preservação da entrada. Usam configuração em memória, sem consultas DNS ou conexões; não incluem IPv6 literal.

## Histórico dos ciclos

Base anterior ao TDD: [`17567023`](https://github.com/surpesaiajin/gatus/commit/175670236bde4e786691fb5858573e9c1e2de544), com a F1 integrada. Execuções de **01/11/2026**, Go **1.27.0 Windows/amd64**, horários UTC−03. Cada linha abaixo corresponde a uma função de teste; os números são de **subtestes**. As variantes de entrada não são ciclos adicionais.

| Etapa | Revisão atual / identificação original | Subtestes PASS / FAIL | Exit | Resultado |
| --- | --- | --- | --- | --- |
| V01 Red | [`bdbbc827`](https://github.com/surpesaiajin/gatus/commit/bdbbc8277bd5098fef9dfc48357de688465c2358) · original `7cb923b8` | 4 / 5 | 1 | Cinco sufixos indevidos são aceitos; quatro controles válidos passam |
| V01 Green | [`06dfe312`](https://github.com/surpesaiajin/gatus/commit/06dfe312399316da57e25c2bbf1862960ffe78c1) · original `0265f258` | 9 / 0 | 0 | Removido o `(.*)` que aceitava texto após a porta |
| V02 Red | [`b40b15fd`](https://github.com/surpesaiajin/gatus/commit/b40b15fd580c94bf5bae80e435f15145be6eb2c9) · original `3ace572e` | 6 / 4 | 1 | `http`, `ip`, `unix` e `tcp7` ainda são aceitos; seis redes permitidas passam |
| V02 Green | [`8699f019`](https://github.com/surpesaiajin/gatus/commit/8699f019380eff29f0043b349390bd1678f8d435) · original `f32e14ff` | 10 / 0 | 0 | Protocolo restrito a `(tcp\|udp)[46]?` |

As duas correções alteram a mesma linha da regex em sequência. Porta vazia continua produzindo erro de conversão, overflow mantém seu erro e a faixa numérica permanece validada. Os testes foram preservados entre Red e Green. Os commits originais foram criados após cada execução, com o conteúdo executado conferido por hash. Suas mensagens foram revisadas posteriormente, mantendo árvores de arquivos, autores, datas e sequência; a tabela relaciona as revisões atuais aos identificadores dos registros originais.

**Refactor:** dispensado com justificativa nos dois ciclos. V01 remove um grupo permissivo; V02 restringe o grupo de protocolo existente. As mudanças não introduzem novos `if`, estado compartilhado ou duplicação no produto. O fluxo de captura, conversão e faixa permanece curto; extrair helpers ou reestruturar o parser não é necessário para esses comportamentos.

## Regressão e CP11

No estado final `8699f019`, a seleção `^TestConfig_` de `./client` passou com **10 funções e 52 subtestes**, exit **0**, sem FAIL ou SKIP. Inclui testes herdados, CP/CB da F1, CTX01 e V01/V02; usa fixtures TLS locais de `testdata/`.

**CP11 passou sem alteração do teste:** `://8.8.8.8:53` é rejeitado porque protocolo vazio está fora das seis redes de V02. É consequência dessa melhoria, não um terceiro ciclo. Nos estados anteriores, inclusive o Green de V01, a regressão ainda falhava apenas em CP11. Os resultados da [parte 1](../f1/README.md) permanecem identificados como anteriores ao TDD.

As [saídas originais transcritas](evidencias/execucoes.txt) reúnem os quatro testes Red/Green e a regressão final válida, com datas, revisões, seleções e contagens. Os registros históricos não são novas execuções.

## Como executar

Na raiz do clone, com Go compatível com `go.mod` e dependências disponíveis:

```powershell
# Nove subtestes de V01; exit 0 no estado final.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_PTOSS2_TDD_V01_DNSResolverSuffix$' ./client

# Dez subtestes de V02; exit 0 no estado final.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_PTOSS2_TDD_V02_DNSResolverNetwork$' ./client

# Recorte completo de configuração: dez funções e 52 subtestes no estado registrado.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_' ./client
```

Os Reds correspondem às revisões indicadas no histórico. O checkout final já contém as correções. O filtro de configuração deixa fora os testes de rede de `client/client_test.go`; não equivale à suíte integral do projeto.

## Cobertura do estado final

Medição válida de **01/11/2026, 14:00:54–14:00:56 UTC−03**, com `-covermode=atomic`: **100,0% de statements de `parseDNSResolver`** e **11,4% do pacote `client`** na seleção `^TestConfig_`. [Relatório por função e pacote](evidencias/client-config.cover-func.txt).

Os percentuais já eram alcançados na parte 1, mesmo com CP11 falhando. Os ciclos corrigem comportamentos sem acrescentar statements; esses percentuais não medem branches, MC/DC ou cobertura global do Gatus.

Para gerar uma nova medição em uma pasta exclusiva, preservando os registros publicados:

```powershell
$saidaTDDI1 = Join-Path ([System.IO.Path]::GetTempPath()) ('gatus-tdd-i1-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $saidaTDDI1 | Out-Null
go test -mod=readonly -count=1 -timeout=2m -covermode=atomic "-coverprofile=$saidaTDDI1/client.cover.out" -run '^TestConfig_' ./client
if ($LASTEXITCODE -ne 0) { throw 'A execução ou a geração do perfil falhou; conferir a saída antes de usar a cobertura.' }
go tool cover "-func=$saidaTDDI1/client.cover.out"
```
