# PTOSS-2 — F4: configuração de consultas DNS

**Responsável:** leonardoabezerra · **Ordem:** caixa-preta → caixa-branca.

**Contexto histórico:** os resultados abaixo registram a parte 1, executada antes das alterações do produto por TDD. O [TDD de I4](../tdd-i4/README.md) registra a evolução posterior; nenhum caso desta parte precisou ser alterado depois dele.

## Objetivo e referência técnica

Testar `(*dns.Config).ValidateAndSetDefault`, que valida e normaliza a configuração `dns` de um endpoint: exige `query-name`, acrescenta o ponto final ao nome e verifica se `query-type` é um tipo conhecido. Os testes usam configuração em memória; nenhuma consulta DNS é enviada.

Execuções de **02/11/2026** na branch `ptoss2-dns-lookup`, HEAD `0437dd85`. Os commits desse HEAD posteriores a `d255d0c` (`v5.38.0`) alteram apenas `client/` e `docs/`: o [método testado](https://github.com/surpesaiajin/gatus/blob/d255d0cd4d69ce818574cfdb3a7e82407e6559d7/config/endpoint/dns/dns.go#L27) é idêntico ao da base (`git hash-object` `47a29bdc`). Os [testes herdados](../../../config/endpoint/dns/dns_test.go) permanecem inalterados. Arquivos novos:

- [Caixa-preta](../../../config/endpoint/dns/dns_ptoss2_test.go): `TestConfig_PTOSS2_F4_BlackBox`, com 11 subtestes CP.
- [Complementação caixa-branca](../../../config/endpoint/dns/dns_ptoss2_whitebox_test.go): `TestConfig_PTOSS2_F4_WhiteBox`, com 4 subtestes CB.

Os casos CP foram projetados a partir da documentação; a análise CB e seus casos vieram depois da execução CP, a partir do código. Houve leitura do código na preparação, então a ordem representa a perspectiva usada para escolher os casos, e não um desconhecimento prévio da implementação.

## Contrato e projeto de caixa-preta

Fontes: a tabela de configuração do [README](../../../README.md) (`endpoints[].dns.query-type`: "Query type (e.g. MX)", padrão `""`; `endpoints[].dns.query-name`: "Query name (e.g. example.com)", padrão `""`), o exemplo da seção *Monitoring an endpoint using DNS queries* (`example.com`, `A`), os comentários públicos do pacote (`QueryType` "like A, AAAA, CNAME...", erros exportados `ErrDNSWithNoQueryName` e `ErrDNSWithInvalidQueryType`) e a orientação da equipe sobre a normalização `example.com` → `example.com.`.

- **R1:** `query-name` é obrigatório; sem ele, `ErrDNSWithNoQueryName`.
- **R2:** `query-type` deve ser um tipo de registro conhecido; tipo desconhecido ou ausente (padrão `""`) gera `ErrDNSWithInvalidQueryType`.
- **R3:** um nome válido é normalizado para FQDN: recebe um único ponto final quando não termina em ponto; nome já terminado em ponto, inclusive a raiz `.`, é mantido.
- **R4 (recorte):** a documentação usa os tipos em maiúsculas e não especifica caixa nem sintaxe do nome. Esses pontos ficaram fora da parte 1 para não congelar em testes um comportamento que o I4 alteraria (V07 e V08).

Classes por condição, variando uma condição inválida por vez e mantendo as demais válidas:

| Condição | Classes válidas | Classes inválidas |
| --- | --- | --- |
| Presença do nome | N1: não vazio | N2: vazio |
| Forma do nome | F1: relativo, sem ponto final; F2: FQDN, com ponto final; F3: raiz `.` | — (sintaxe fora do recorte, ver V08) |
| Presença do tipo | P1: informado | P2: vazio, o padrão do README |
| Valor do tipo | T1: tipo conhecido, em maiúsculas | T2: texto que não é tipo |

Valores-limite: comprimento do nome 0 (CP05) e 1 (CP03, CP04); comprimento do tipo 0 (CP09) e 1 (`A`, CP01); `AAAAA` (CP11) como vizinho de um valor válido, `AAAA` com um caractere a mais.

**OK** exige erro nulo, `QueryName` igual ao esperado e `QueryType` inalterado. **ERRO** exige `errors.Is` com o erro exportado correspondente: os erros exportados e documentados fazem parte do contrato público. CP não verifica o estado da configuração depois de um erro.

| Caso | `query-name` | `query-type` | Classe/técnica | Esperado | Resultado |
| --- | --- | --- | --- | --- | --- |
| CP01 | `example.com` | `A` | N1, F1, P1, T1; exemplo do README; tipo de comprimento 1 | OK, `example.com.` | PASS |
| CP02 | `example.com.` | `A` | F2; equivalência | OK, inalterado | PASS |
| CP03 | `.` | `A` | F3; nome de 1 caractere terminado em ponto | OK, `.` | PASS |
| CP04 | `a` | `A` | F1; menor nome não vazio sem ponto | OK, `a.` | PASS |
| CP05 | `""` | `A` | N2; comprimento 0 | `ErrDNSWithNoQueryName` | PASS |
| CP06 | `example.com` | `AAAA` | T1; registro IPv6 | OK, `example.com.` | PASS |
| CP07 | `example.com` | `MX` | T1; exemplo do README | OK, `example.com.` | PASS |
| CP08 | `example.com` | `CNAME` | T1; citado no comentário do campo | OK, `example.com.` | PASS |
| CP09 | `example.com` | `""` | P2; comprimento 0 | `ErrDNSWithInvalidQueryType` | PASS |
| CP10 | `example.com` | `XYZ` | T2; equivalência | `ErrDNSWithInvalidQueryType` | PASS |
| CP11 | `example.com` | `AAAAA` | T2; vizinho de valor válido | `ErrDNSWithInvalidQueryType` | PASS |

Nenhum caso CP revelou falha.

## Decisões, caminhos e MC/DC

Mapeamento manual do código da base:

| ID / linha | Decisão | Representante verdadeiro | Representante falso |
| --- | --- | --- | --- |
| D1 / 28 | `len(d.QueryName) == 0` | CP05 | CP01–CP04, CP06–CP11 |
| D2 / 31 | `!strings.HasSuffix(d.QueryName, ".")` | CP01, CP04, CP06–CP11 | CP02, CP03 |
| D3 / 34 | `!ok`, tipo ausente de `dns.StringToType` | CP09–CP11 | CP01–CP04, CP06–CP08 |

Com CP, os **6/6 resultados** das três decisões têm representantes. As três decisões têm **uma condição cada**, então MC/DC coincide com a cobertura de decisões: o par V/F de cada decisão demonstra a influência independente de sua única condição. O método não tem decisão composta. A conjunção interna de `strings.HasSuffix` (`len(s) >= len(suffix) && ...`) pertence à biblioteca padrão; neste método o primeiro operando é sempre verdadeiro após D1 falso, então um par MC/DC para ele seria inviável aqui e fica fora do escopo.

Os resultados por decisão não bastaram para guiar a complementação. A análise de **caminhos** mostrou cinco sequências viáveis:

| Caminho | D1 | D2 | D3 | Retorno | Casos CP | Casos CB |
| --- | --- | --- | --- | --- | --- | --- |
| P1 | V | — | — | `ErrDNSWithNoQueryName` | CP05 | CB01 |
| P2 | F | V | V | `ErrDNSWithInvalidQueryType` | CP09–CP11 | CB02 |
| P3 | F | V | F | `nil` | CP01, CP04, CP06–CP08 | CB04 |
| P4 | F | F | V | `ErrDNSWithInvalidQueryType` | — | CB03 |
| P5 | F | F | F | `nil` | CP02, CP03 | — |

CP percorreu 4/5 caminhos; **P4, FQDN com tipo inválido, ficou sem caso**. O perfil `count` de CP mostra o efeito sem acusá-lo: D2 foi avaliada 11 vezes e o acréscimo do ponto rodou 9, logo D2 falso ocorreu apenas 2 vezes, ambas com tipo válido. O `if` sem `else` não gera bloco para o resultado falso, então 100% de statements não revela esse caminho.

## Complementação de caixa-branca

Em todos os CB, `errors.Is` verifica o erro esperado (ou `nil`), e o teste confere `QueryName` e `QueryType` **também nos caminhos de erro**, o que CP não fazia.

| Caso | Entrada | Decisões / motivo estrutural | Oráculo específico | Resultado |
| --- | --- | --- | --- | --- |
| CB01 | `""`, `XYZ` | D1=V retorna antes de D2 e D3 | `ErrDNSWithNoQueryName` tem precedência; nome continua `""` | PASS |
| CB02 | `example.com`, `XYZ` | D2=V executa antes de D3=V | erro de tipo, mas o nome **já foi alterado** para `example.com.` | PASS |
| CB03 | `example.com.`, `XYZ` | caminho P4, ausente em CP | erro de tipo; nome mantido | PASS |
| CB04 | `example.com`, `ANY` | D3 consulta a tabela da biblioteca | meta-tipo de consulta aceito | PASS |

CB acrescentou três entradas novas (CB01, CB03, CB04) e um oráculo de efeito colateral (CB02) sobre uma entrada equivalente à de CP10. Nenhum CB revelou falha: os resultados confirmam o comportamento atual e o tornam explícito.

**Observações da análise estrutural, não convertidas em testes que falham:**

- **O1 — classe válida de tipos:** D3 aceita qualquer chave de `dns.StringToType`, o inverso de `dns.TypeToString`. Isso inclui meta-tipos (`ANY`, `AXFR`, `IXFR`) e duas chaves com letras mistas, `None` (tipo 0) e `Reserved` (65535). Esse comportamento acidental não foi congelado em teste; o efeito do V07 sobre ele está registrado no I4.
- **O2 — ponto final escapado:** D2 examina apenas o último caractere. `example\.` (ponto escapado, que não é FQDN segundo `dns.IsFqdn`) é mantido sem ponto final. Caso raro do formato de apresentação; registrado para o grupo, sem teste.
- **O3 — integração com `client.QueryDNS`:** a normalização incondicional impede a conversão automática de IP em consultas `PTR` (#1261). `1.0.0.1` vira `1.0.0.1.`, que `net.ParseIP` rejeita. Uma sonda temporária em `./client` (não entregue; removida após a execução de 02/11) obteve `normalized="1.0.0.1." err=invalid IP: 1.0.0.1.`, sem consulta de rede. A falha fica no cliente, fora do recorte da F4. Pela regra da equipe, foi registrada para discussão e não foi corrigida.

## Resultados e cobertura antes/depois

Execuções de **02/11/2026**, Go **1.27.0 linux/amd64** na imagem oficial `golang:1.27.0` (Docker Desktop em Windows 11), horários do host UTC−03, zero SKIP. O pacote `config/endpoint/dns` contém apenas este método, então a cobertura do pacote é igual à do método.

| Etapa / seleção | Horário | Funções PASS/FAIL | Subtestes PASS/FAIL | Exit | Statements | Decisões (manual) | Caminhos (manual) |
| --- | --- | --- | --- | --- | --- | --- | --- |
| [Antes, somente herdados](evidencias/antes/dns.cover-func.txt) | 12:19:58–12:20:07 | 2/0 | 0/0 | 0 | 85,7% (6/7) | 4/6 | 2/5 |
| [Herdados + CP](evidencias/cp/dns.cover-func.txt) | 12:21:14–12:21:15 | 3/0 | 11/0 | 0 | 100% (7/7) | 6/6 | 4/5 |
| [Herdados + CP + CB](evidencias/cp-cb/dns.cover-func.txt) | 12:21:37–12:21:38 | 4/0 | 15/0 | 0 | 100% (7/7) | 6/6 | 5/5 |
| [Somente CB](evidencias/cb/dns.cover-func.txt) | 14:04:07–14:04:10 | 1/0 | 4/0 | 0 | 100% (7/7) | 6/6 | 4/5 |

Antes, os dois testes herdados exercitavam apenas caminhos de erro (P1 e P2): o `return nil` (linha 37) nunca rodava, e D2 falso e D3 falso não tinham representante. CP elevou statements a 100% e completou os resultados das decisões; CB não alterou o percentual, mas fechou o caminho P4 e acrescentou oráculos de estado. A seleção somente CB, executada depois sobre o mesmo produto da base e sem o arquivo do TDD na árvore, mede outro recorte: os quatro casos também atingem 100% de statements e 6/6 resultados, mas não percorrem P5, que só CP02 e CP03 exercitam. Todas as seleções usam `-covermode=count`. Os perfis completos ([antes](evidencias/antes/dns.cover.out), [CP](evidencias/cp/dns.cover.out), [CP+CB](evidencias/cp-cb/dns.cover.out), [somente CB](evidencias/cb/dns.cover.out)) mostram os contadores por bloco. As [saídas integrais](evidencias/execucoes.txt) incluem horários, HEAD, estado da árvore e o `git hash-object` de cada arquivo executado.

## Como executar

Na raiz deste clone, com Go compatível com `go.mod`:

```powershell
# CP01–CP11 e CB01–CB04: código 0 no estado registrado.
go test -mod=readonly -count=1 -v -run '^TestConfig_PTOSS2_F4_(BlackBox|WhiteBox)$' ./config/endpoint/dns

# Pacote inteiro, com perfil de cobertura em pasta temporária.
$saidaF4 = Join-Path ([System.IO.Path]::GetTempPath()) ('gatus-f4-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $saidaF4 | Out-Null
go test -mod=readonly -count=1 -covermode=count "-coverprofile=$saidaF4/dns.cover.out" ./config/endpoint/dns
go tool cover "-func=$saidaF4/dns.cover.out"
```

Sem Go instalado, o mesmo comando roda pela imagem usada nas execuções (as dependências ficam em volumes nomeados):

```powershell
docker run --rm -v gatus-gomod:/go/pkg/mod -v gatus-gocache:/root/.cache/go-build -v "${PWD}:/src" -w /src golang:1.27.0 go test -mod=readonly -count=1 -v ./config/endpoint/dns
```

No checkout atual, que já contém o I4, o produto tem uma decisão a mais e a seleção do pacote inclui também V07 e V08. Para reproduzir as linhas acima, use o `dns.go` da base (`git show d255d0cd:config/endpoint/dns/dns.go`); para a linha "Antes", restrinja a execução aos herdados com `-run '^TestConfig_ValidateAndSetDefault(sWithInvalidDNSQueryType)?$'`; para "Somente CB", use `-run '^TestConfig_PTOSS2_F4_WhiteBox$'`.

---

**```Nota: este README.md foi escrito com uso de Inteligência Artificial para fins de documentação```**