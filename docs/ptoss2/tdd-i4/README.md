# PTOSS-2 — I4: TDD da configuração de consultas DNS

**Responsável:** leonardoabezerra · **Ciclos:** V07 e V08, na numeração da equipe.

## Comportamentos desenvolvidos

O [método](../../../config/endpoint/dns/dns.go) `(*dns.Config).ValidateAndSetDefault` é chamado por `Endpoint.ValidateAndSetDefaults` para todo endpoint com bloco `dns`. Os dois ciclos evoluem comportamentos distintos:

- **V07:** aceitar `query-type` em minúsculas ou com letras misturadas e armazená-lo em maiúsculas (`a` → `A`, `mX` → `MX`). Tipos desconhecidos continuam inválidos. O valor armazenado é o que `client.QueryDNS` procura em `dns.StringToType`, então guardá-lo em maiúsculas é necessário para a consulta usar o tipo pedido.
- **V08:** detectar já na configuração os nomes que a biblioteca `miekg/dns` não aceita, como `a..example`, um rótulo com mais de 63 caracteres ou um ponto inicial, com o novo erro `ErrDNSWithInvalidQueryName`. Nomes válidos, a raiz `.`, nomes de serviço com underscore (`_sip._tcp.example.com`) e o IPv4 usado em consultas `PTR` são preservados.

Os [testes](../../../config/endpoint/dns/dns_ptoss2_tdd_test.go) usam configuração em memória, sem consultas DNS ou conexões. Cada função corresponde a um comportamento; as variantes de entrada de cada uma não são ciclos adicionais.

## Histórico dos ciclos

Base anterior ao TDD: HEAD `0437dd85` com a [F4](../f4/README.md) na árvore de trabalho e `dns.go` idêntico a `d255d0c` (`47a29bdc`). Go **1.27.0 linux/amd64** (imagem oficial `golang:1.27.0`), horários do host UTC−03, zero SKIP. Nenhum commit foi criado durante os ciclos; o `git hash-object` dos arquivos registrado em cada execução identifica o conteúdo executado. Os números abaixo são de **subtestes** da função do ciclo.

| Etapa | Horário (02/11/2026) | Conteúdo executado (`git hash-object`) | Subtestes PASS / FAIL | Exit | Resultado |
| --- | --- | --- | --- | --- | --- |
| V07 Red | 14:04:11–14:04:12 | teste `68625f25`; `dns.go` `47a29bdc` (base) | 3 / 4 | 1 | `a`, `mX`, `aaaa` e `Cname` recebem `invalid query type`; `A`, `xyz` e `aAaAa` passam |
| V07 Green | 14:04:12–14:04:14 | teste `68625f25`; `dns.go` `af16018e` | 7 / 0 | 0 | uma linha: `d.QueryType = strings.ToUpper(d.QueryType)` antes de D3 |
| V08 Red | 14:04:36–14:04:37 | teste `dc6dd040`; `dns.go` `94696029` | 6 / 3 | 1 | `a..example`, `.example.com` e o rótulo de 64 caracteres retornam erro nulo; seis controles passam |
| V08 Green | 14:04:37–14:04:38 | teste `dc6dd040`; `dns.go` `e91cb56e` | 9 / 0 | 0 | `dns.IsDomainName` após a normalização do nome |

Os testes não mudaram entre Red e Green de cada ciclo; o arquivo de V08 apenas acrescentou a segunda função ao de V07. No Red de V08, `dns.go` declara `ErrDNSWithInvalidQueryName` **sem usá-lo**: sem essa declaração o pacote não compilaria e a falha seria de compilação, não de comportamento. A declaração não altera nenhuma saída do método; o Green acrescenta a decisão que retorna esse erro.

Os ciclos foram executados primeiro às 12:24–12:27 com os identificadores provisórios V01/V02. Para seguir a numeração da equipe, as funções foram renomeadas para V07/V08 e a sequência foi reexecutada a partir das cópias de cada etapa, com os mesmos testes e produtos; as contagens não mudaram. O [registro](evidencias/execucoes.txt) é o dessa reexecução.

Diferença final do produto em relação à base:

```diff
+	// ErrDNSWithInvalidQueryName is the error with which gatus will panic if a dns is configured with invalid query name
+	ErrDNSWithInvalidQueryName = errors.New("invalid query name in the DNS configuration")
 ...
 	if !strings.HasSuffix(d.QueryName, ".") {
 		d.QueryName += "."
 	}
+	if _, ok := dns.IsDomainName(d.QueryName); !ok {
+		return ErrDNSWithInvalidQueryName
+	}
+	d.QueryType = strings.ToUpper(d.QueryType)
 	if _, ok := dns.StringToType[d.QueryType]; !ok {
```

**Refactor:** dispensado com justificativa nos dois ciclos. V07 acrescenta uma atribuição na sequência existente; V08 acrescenta uma decisão de uma condição no mesmo padrão `if _, ok := ...; !ok` já usado em D3. Nenhum ciclo criou duplicação, estado compartilhado ou ramo aninhado; o método segue linear, com quatro verificações em ordem. Trocar o `HasSuffix` manual por `dns.Fqdn` foi considerado e rejeitado como refatoração: `dns.Fqdn` trata o ponto final escapado de outro modo (observação O2 da F4), então mudaria o comportamento. Nos testes, as duas funções têm tabelas e oráculos diferentes, e extrair um helper comum ocultaria o que cada uma verifica.

## Decisões de projeto e consequências

- **Ordem das verificações:** o nome é validado logo após a normalização e antes do tipo. Erros de nome continuam tendo precedência sobre erros de tipo, como D1 já tinha na base.
- **Biblioteca existente:** a validação delega a `dns.IsDomainName`, como pedido. A função é liberal por projeto: aceita espaços e `*`, por exemplo, e verifica rótulos vazios, ponto inicial, rótulos de até 63 octetos e o limite total do formato de transmissão. O limite exato do comprimento total segue a contagem da biblioteca e não foi testado como valor-limite.
- **Valor-limite em V08:** o rótulo de 63 caracteres é aceito e o de 64 é rejeitado.
- **Consequência de V07:** `None` e `Reserved`, as duas chaves de `dns.StringToType` com letras mistas (observação O1 da F4), deixam de ser aceitas, porque `strings.ToUpper` produz `NONE` e `RESERVED`. Nenhuma das duas é um tipo de consulta útil (tipo 0 e 65535). A mudança foi registrada, sem teste que a congele.
- **Erro de tipo:** em caminho de erro, o tipo fica armazenado já convertido para maiúsculas. Esse estado não é especificado e não é verificado por V07.
- **`PTR` com IPv4:** V08 mantém `8.8.8.8` aceito na configuração. A falha O3 da F4, em que o cliente recebe `8.8.8.8.` e não converte o IP, continua existindo no cliente; V08 não a agrava nem a corrige.

## Regressão

Depois de cada Green, a seleção do pacote `./config/endpoint/dns` passou integralmente: 5 funções e 22 subtestes após V07; 6 funções e 31 subtestes após V08. Os 15 subtestes CP/CB da F4 continuaram passando **sem alteração**: caixa e sintaxe do nome tinham ficado deliberadamente fora da parte 1. `gofmt -l` e `go vet` não apontaram problemas no pacote.

`go test ./config/...` terminou com exit 0 após V08. Após V07, a primeira execução (14:04:15–14:04:35) terminou com exit 1 por `TestSuite_ExecuteWithAlwaysRunEndpoints`, de `config/suite`: o teste herdado faz requisições HTTPS reais a `example.org`, a primeira demorou cerca de 13 s e falhou, e o resultado teve 2 endpoints em vez de 3. O pacote não usa configuração DNS. A repetição no mesmo conteúdo (14:05:24–14:05:34) terminou com exit 0 em todos os pacotes; as duas execuções estão no registro.

As [saídas integrais](evidencias/execucoes.txt) reúnem Red, Green, as regressões e a medição final, com horários, HEAD, estado da árvore, hashes e um resumo das contagens.

## Cobertura do estado final

Medição de **02/11/2026, 14:04:49–14:04:50 UTC−03**, `-covermode=count`, pacote inteiro: 6 funções e 31 subtestes PASS, **100,0% de statements** de `ValidateAndSetDefault` (10/10). [Relatório por função](evidencias/dns.cover-func.txt) e [perfil com contadores](evidencias/dns.cover.out).

| ID | Decisão no estado final | Representante verdadeiro | Representante falso |
| --- | --- | --- | --- |
| D1 | `len(d.QueryName) == 0` | CP05 | demais |
| D2 | `!strings.HasSuffix(d.QueryName, ".")` | CP01 | CP02 |
| DN | `!ok` de `dns.IsDomainName` (nova) | V08 `empty_label_between_dots` | V08 `root_is_kept` |
| D3 | `!ok` de `dns.StringToType` | CP10, V07 `unknown_lowercase_type_is_rejected` | CP01, V07 `lowercase_a` |

As quatro decisões têm uma condição cada; os 8/8 resultados têm representantes, e MC/DC coincide com a cobertura de decisões. Os percentuais não medem a cobertura global do Gatus.

## Como executar

Na raiz do clone, com Go compatível com `go.mod`:

```powershell
# Sete subtestes de V07; exit 0 no estado final.
go test -mod=readonly -count=1 -v -run '^TestConfig_PTOSS2_TDD_V07_QueryTypeCase$' ./config/endpoint/dns

# Nove subtestes de V08; exit 0 no estado final.
go test -mod=readonly -count=1 -v -run '^TestConfig_PTOSS2_TDD_V08_QueryNameSyntax$' ./config/endpoint/dns

# Pacote inteiro (herdados, F4, V07 e V08) e regressão da configuração.
go test -mod=readonly -count=1 ./config/endpoint/dns
go test -mod=readonly -count=1 ./config/...
```

Para executar pela imagem Docker usada nas medições, veja a seção *Como executar* da [F4](../f4/README.md#como-executar). Os Reds correspondem aos conteúdos identificados pelos hashes do histórico; o checkout final já contém as duas implementações. `config/suite` depende de acesso à internet.

---

**```Nota: este README.md foi escrito com uso de Inteligência Artificial para fins de documentação```**