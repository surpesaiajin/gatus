# PTOSS-2 — F1: parser do resolvedor DNS

**Responsável:** surpesaiajin · **Ordem:** caixa-preta → caixa-branca.

## Objetivo e referência técnica

Testar `(*client.Config).parseDNSResolver`, que separa protocolo, host e porta de `Config.DNSResolver` ou retorna erro. Os testes verificam configuração em memória, sem consultas DNS ou conexões ao resolvedor.

Base das execuções: **`d255d0cd4d69ce818574cfdb3a7e82407e6559d7` (`v5.38.0`)**, branch `ptoss2-f1-dns-resolver`. O [método](../../../client/config.go#L155) e os [testes herdados](../../../client/config_test.go) permanecem inalterados. Arquivos novos:

- [Caixa-preta e CTX01](../../../client/config_dns_resolver_ptoss2_test.go): duas funções principais, sendo uma com 17 subtestes CP.
- [Complementação caixa-branca](../../../client/config_dns_resolver_ptoss2_whitebox_test.go): uma função principal com cinco subtestes CB.

Os casos CP foram projetados antes de sua implementação; a análise CB e seus casos foram registrados depois da execução CP. Já existia contato com o código na preparação: a ordem representa a perspectiva usada para selecionar os casos, sem alegar desconhecimento prévio da implementação.

## Contrato e projeto de caixa-preta

O [README do Gatus](../../../README.md#client-configuration) documenta `{proto}://{host}:{port}` e o exemplo `tcp://8.8.8.8:53`. Adotamos estas expectativas:

- **R1:** entrada válida retorna os três componentes exatamente, com porta como string, sem erro.
- **R2:** formato incompleto ou componente obrigatório ausente produz erro na chamada direta ao parser. É uma interpretação operacional do formato; o README não enumera todos os erros.
- **R3:** porta decimal entre **1 e 65535**, contrato da equipe preservado a partir da leitura anterior do código. Esse intervalo não é apresentado como promessa explícita do README.
- **R4:** TCP, UDP e `localhost` são controles do recorte; não comprovam suporte ou validação completa de todos os protocolos e nomes.
- **R5:** opção vazia significa ausência de override no cliente; isso não exige que o parser chamado diretamente aceite string vazia.

Classes por condição, mantendo as demais válidas ao variar uma condição inválida:

| Condição | Classes válidas no recorte | Classes inválidas no recorte |
| --- | --- | --- |
| Entrada direta | E01: não vazia | E02: vazia |
| Separador de protocolo | F01: `://` | F02: ausente/incorreto |
| Protocolo | N01: `tcp`; N02: `udp` | N03: ausente |
| Host | H01: IPv4 de controle; H02: `localhost` | H03: ausente |
| Presença de porta | A01: `:` e valor | A02: ambos ausentes; A03: valor vazio |
| Representação da porta | T01: algarismos sem sinal | T02: letras; T03: sinal negativo |
| Valor decimal | P01: `[1,65535]` | P02: abaixo, representado por `0`; P03: acima, por `65536` |

Não é um produto cartesiano de todas as strings. **OK** abaixo exige erro nulo, resultado não nulo e campos iguais aos componentes da entrada. **ERRO** exige erro não nulo, sem fixar mensagem ou exigir resultado nulo nesta perspectiva. Todos verificam preservação de `DNSResolver`.

Resultados reais de **30/10/2026**:

| Caso | Entrada | Classe/técnica principal | Esperado | Resultado |
| --- | --- | --- | --- | --- |
| CP01 | `tcp://8.8.8.8:53` | Classes válidas; exemplo documentado | OK | PASS |
| CP02 | `udp://8.8.8.8:53` | N02; equivalência | OK | PASS |
| CP03 | `tcp://localhost:53` | H02; equivalência | OK | PASS |
| CP04 | `tcp://8.8.8.8:0` | P02; mínimo−1 | ERRO | PASS |
| CP05 | `tcp://8.8.8.8:1` | P01; mínimo | OK | PASS |
| CP06 | `tcp://8.8.8.8:2` | P01; mínimo+1 | OK | PASS |
| CP07 | `tcp://8.8.8.8:65534` | P01; máximo−1 | OK | PASS |
| CP08 | `tcp://8.8.8.8:65535` | P01; máximo | OK | PASS |
| CP09 | `tcp://8.8.8.8:65536` | P03; máximo+1 | ERRO | PASS |
| CP10 | `tcp:8.8.8.8:53` | F02; equivalência inválida | ERRO | PASS |
| CP11 | `://8.8.8.8:53` | N03; equivalência inválida | ERRO | **FAIL: erro nulo** |
| CP12 | `tcp://:53` | H03; equivalência inválida | ERRO | PASS |
| CP13 | `tcp://8.8.8.8` | A02; equivalência inválida | ERRO | PASS |
| CP14 | `tcp://8.8.8.8:` | A03; equivalência inválida | ERRO | PASS |
| CP15 | `tcp://8.8.8.8:abc` | T02; equivalência inválida | ERRO | PASS |
| CP16 | `tcp://8.8.8.8:-1` | T03; sintaxe, não prova do limite numérico | ERRO | PASS |
| CP17 | `""` | E02; ausência/presença | ERRO direto no parser | PASS |

**CTX01, separado:** `Config{}` passa em `ValidateAndSetDefaults`, mantém `DNSResolver` vazio e `HasCustomDNSResolver()` falso. Resultado PASS. É um controle da configuração pública, não um 18º caso do parser. O01 (saída válida com campos exatos) e O02 (erro para entrada inválida) são os oráculos associados, não classes adicionais de entrada.

**Falha conhecida CP11:** a captura de protocolo usa `.*` e aceita protocolo vazio. A expectativa R2 foi preservada por decisão do responsável. O caso continua ativo, sem skip ou correção do produto. A parte 1 registra essa divergência; sua conclusão não depende de corrigir o produto. As seleções que incluem CP11 retornam código **1**.

## Decisões, laço e MC/DC

Mapeamento manual do código da base, apoiado nos testes executados:

| ID / linha | Decisão | Representante verdadeiro | Representante falso |
| --- | --- | --- | --- |
| D1 / 158 | `len(matches) == 0` | CP10: erro de formato | CP05: continua |
| D2 / 163 | `A && B`, A: `i != 0`, B: `k != ""` | CP05, índice 1 | CP05, índices 0 e 2 |
| D3 / 168 | `err != nil`, após `Atoi` | CP14: porta capturada vazia | CP05: converte `1` |
| D4 / 171 | `C \|\| D`, C: `port < 1`, D: `port > 65535` | CP04 e CP09 | CP05 e CP08 |
| L1 / 162 | Continuação/término do `range` | CP05: seis iterações | CP05: término após a sexta |

São **8/8 resultados dos quatro `if` explícitos** com representantes, mais entrada/término do laço. Essa é uma contagem manual delimitada, não uma métrica instrumentada de todas as branches do pacote ou do compilador.

Tabelas lógicas das decisões compostas (**NE**: operando não avaliado por curto-circuito; seu valor lógico é deduzido):

| D2: A | B lógico | Resultado | Avaliação/viabilidade |
| --- | --- | --- | --- |
| F | F | F | Índice 0, nome vazio; B é NE |
| F | V | F | Inviável: índice 0 sempre tem nome vazio |
| V | F | F | Índices 2 ou 5, grupos sem nome |
| V | V | V | Índices 1, 3 ou 4, grupos nomeados |

`SubexpNames()` é fixo: `["", "proto", "", "host", "port", ""]`. Para **B**, o par V/F → V/V, em CP05 nos índices 2/1, muda a decisão mantendo A=V. Para **A**, seria necessário F/V → V/V; a primeira combinação é inalcançável. F/F → V/V altera duas condições e não serve. São pares de avaliações em iterações, não entradas que mudam os nomes dos grupos. Toda chamada que alcança o laço executa seis iterações; zero ou uma após alcançá-lo são inviáveis. Retornar em D1 não é executar o laço zero vezes após sua entrada.

| D4: C | D lógico | Resultado | Representante/viabilidade |
| --- | --- | --- | --- |
| F | F | F | CP05 (`1`) ou CP08 (`65535`) |
| F | V | V | CP09 (`65536`) |
| V | F | V | CP04 (`0`); D é NE |
| V | V | V | Inviável para um mesmo inteiro |

Pares: **C: CP04/CP05**, mantendo D lógico falso; **D: CP08/CP09**, mantendo C falso. D4 tem pares para **2/2 condições**. No conjunto D2+D4, são **3 condições com par entre 4**, com uma inviável. **Não há demonstração de MC/DC integral do método.** Os perfis Go não medem operandos; a justificativa de independência e inviabilidade é manual.

## Complementação de caixa-branca

A perspectiva estrutural acrescentou a identificação específica dos erros e um overflow, mesmo com statements já em 100%. Em todos os CB, `errors.Is` verifica o erro indicado, o resultado deve ser nulo e a entrada deve permanecer igual.

| Caso | Entrada/relação com CP | Oráculo específico | Resultado em 30/10 |
| --- | --- | --- | --- |
| CB01 | Igual a CP10 | `ErrInvalidDNSResolver` | PASS |
| CB02 | Igual a CP14 | `strconv.ErrSyntax` | PASS |
| CB03 | `tcp://8.8.8.8:` + 100 algarismos `9` | `strconv.ErrRange` de conversão, antes de D4 | PASS |
| CB04 | Igual a CP04 | `ErrInvalidDNSResolverPort` | PASS |
| CB05 | Igual a CP09 | `ErrInvalidDNSResolverPort` | PASS |

Quatro casos reforçam oráculos de entradas existentes; CB03 é a única entrada nova. Não são cinco novas classes funcionais. Na regex atual, CP15/CP16 também chegam a `Atoi` com porta capturada vazia; não demonstram, por si, erro de faixa. O tipo de erro e o resultado nulo foram derivados da implementação na etapa CB, sem atribuí-los retrospectivamente ao projeto CP.

## Resultados e cobertura antes/depois

**Execuções de 30/10/2026**, Go **1.27.0 windows/amd64**, horários UTC−03. Em todas as seleções abaixo houve zero SKIP. Os links da tabela abrem os relatórios de cobertura por função das três etapas.

| Etapa / seleção | Horário | Funções PASS/FAIL | Subtestes PASS/FAIL | Exit | Statements F1 / pacote `client` |
| --- | --- | --- | --- | --- | --- |
| [Antes dos novos testes](evidencias/antes/client-config.cover-func.txt), `^TestConfig_` | 23:07:59–23:08:09 | 5/0 | 11/0 | 0 | 92,9% / 11,2% |
| [Herdados + CP + CTX01](evidencias/cp/client-config.cover-func.txt), `^TestConfig_` | 23:10:03–23:10:06 | 6/1 | 27/1 | 1 | 100% / 11,4% |
| [Herdados + CP + CTX01 + CB](evidencias/cp-cb/client-config.cover-func.txt), `^TestConfig_` | 23:43:29–23:43:30 | 7/1 | 32/1 | 1 | 100% / 11,4% |
| CB isolado, `^TestConfig_PTOSS2_F1_WhiteBox$` | 23:43:25–23:43:28 | 1/0 | 5/0 | 0 | 92,9% / 2,6% |

As três primeiras linhas usam o mesmo filtro sobre o mesmo código de produção, com acréscimo dos testes. A falha é somente CP11; os cinco testes principais e onze subtestes herdados continuam PASS. A linha CB isolado tem outro recorte e não indica perda de cobertura. A contagem de funções inclui CTX01 e funções agregadoras, sem contá-las novamente como subtestes.

A cobertura de **statements da F1 aumentou de 92,9% para 100% com CP**; CB não elevou esse percentual. O pacote passou de 11,2% a 11,4% no recorte conjunto, sem significar cobertura global do projeto. Os perfis anteriores usam `mode: set` e os CB usam `mode: count`: é possível comparar statements atingidos, mas não equiparar contadores entre os modos. Cobertura completa de statements coexistiu com CP11 falhando.

## Como executar

Na raiz deste clone, com Go compatível com `go.mod` e dependências disponíveis. As seleções foram examinadas: usam configuração em memória e fixtures TLS locais de `testdata/`, na raiz do repositório. Os demais testes de rede de `./client` ficam fora do filtro. A obtenção de dependências, se necessária, é separada da ausência de rede durante os testes selecionados.

```powershell
# CP01–CP17 + CTX01: código 1 esperado enquanto CP11 estiver preservado.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_PTOSS2_F1_(BlackBox|EmptyOverrideUsesDefaultResolver)$' ./client

# CB01–CB05: código 0 observado na base registrada.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_PTOSS2_F1_WhiteBox$' ./client

# Herdados + CP + CTX01 + CB: código 1 pela falha conhecida CP11.
go test -mod=readonly -count=1 -timeout=2m -run '^TestConfig_' ./client
```

Para repetir apenas a seleção herdada no checkout atual, use `-run '^TestConfig_(getHTTPClient.*|ValidateAndSetDefaults_withCustomDNSResolver|TlsIsValid)$'`. O filtro histórico `^TestConfig_` executava apenas os herdados antes da criação dos novos arquivos; hoje inclui os novos. Não o chamar de baseline inalterado no checkout atual.

Exemplo de nova coleta em PowerShell, em pasta temporária exclusiva, preservando os registros históricos:

```powershell
$saidaF1 = Join-Path ([System.IO.Path]::GetTempPath()) ('gatus-f1-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $saidaF1 | Out-Null
go test -mod=readonly -count=1 -timeout=2m -json -covermode=count "-coverprofile=$saidaF1/client-config.cover.out" -run '^TestConfig_' ./client > "$saidaF1/client-config.test.jsonl"
$codigoTestesF1 = $LASTEXITCODE
go tool cover "-func=$saidaF1/client-config.cover.out" > "$saidaF1/client-config.cover-func.txt"
go tool cover "-html=$saidaF1/client-config.cover.out" "-o=$saidaF1/client-config.cover.html"
Write-Output "Saida: $saidaF1; codigo dos testes: $codigoTestesF1"
```

Código 1 não deve ser mascarado como aprovação. Conferir no log que a falha continua sendo CP11; qualquer outra falha requer análise. Os comandos acima são instruções de reprodução; os resultados da tabela correspondem às execuções de 30/10.
