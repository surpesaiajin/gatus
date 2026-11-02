# PTOSS-2 — F5: testes de `OAuth2Config.isValid`

**Responsável:** gccintra · **Base:** [`17567023`](https://github.com/surpesaiajin/gatus/commit/175670236bde4e786691fb5858573e9c1e2de544) (`ptoss2`, sobre `d255d0c` / v5.38.0) · **Execuções:** 02/11/2026, Go 1.26.3 darwin/arm64, horários UTC−03.

## Comportamento testado

[`OAuth2Config.isValid`](../../../client/config.go#L202) informa se a configuração OAuth2 (fluxo *client credentials*) tem os quatro dados necessários: `token-url`, `client-id`, `client-secret` e `scopes`. É chamado por [`Config.ValidateAndSetDefaults`](../../../client/config.go#L135), que devolve `ErrInvalidClientOAuth2Config` quando há bloco `oauth2` incompleto. Não se verifica se a credencial é aceita por um provedor real.

```go
return len(c.TokenURL) > 0 && len(c.ClientID) > 0 && len(c.ClientSecret) > 0 && len(c.Scopes) > 0
```

Ordem pedida: **caixa-branca primeiro**, depois caixa-preta para encontrar lacunas.

## 1. Caixa-branca — decisões e MC/DC

[Testes](../../../client/config_oauth2_ptoss2_whitebox_test.go). Condições: **A** `len(TokenURL)>0`, **B** `len(ClientID)>0`, **C** `len(ClientSecret)>0`, **D** `len(Scopes)>0`. Decisão `A && B && C && D`; MC/DC exige N+1 = 5 casos.

| Caso | A | B | C | D | Resultado | Par MC/DC |
| --- | --- | --- | --- | --- | --- | --- |
| CB01 | V | V | V | V | `true` | base |
| CB02 | **F** | V | V | V | `false` | (CB01, CB02) mostra A |
| CB03 | V | **F** | V | V | `false` | (CB01, CB03) mostra B |
| CB04 | V | V | **F** | V | `false` | (CB01, CB04) mostra C |
| CB05 | V | V | V | **F** | `false` | (CB01, CB05) mostra D |

Com o curto-circuito do `&&` em Go, em CB02 as condições B–D não são avaliadas (e assim por diante); os pares continuam válidos como MC/DC de causa única com mascaramento, porque só a condição em foco muda entre os dois casos.

Decisão chamadora em `ValidateAndSetDefaults`: **P** `HasOAuth2Config()` e **Q** `!isValid()`, decisão `P && Q`.

| Caso | P | Q | Resultado | Par MC/DC |
| --- | --- | --- | --- | --- |
| CB06 | **F** | (não avaliada) | sem erro | (CB06, CB08) mostra P |
| CB07 | V | **F** | sem erro | (CB07, CB08) mostra Q |
| CB08 | V | V | `ErrInvalidClientOAuth2Config` | — |

Cada caso também confere que o método não altera a configuração. Resultado: **2 funções, 8 subtestes PASS**; `isValid` passou de **0,0% para 100,0%** de statements.

## 2. Caixa-preta — classes e limites

[Testes](../../../client/config_oauth2_ptoss2_test.go). Classes por campo de texto: ausente (`""`) / preenchido, com o limite inferior de 1 caractere. Para `scopes`: ausente (`nil`), vazia (`[]`), vazia com capacidade, 1 elemento (limite) e vários elementos. Cada caso é verificado em `isValid` e na API pública `ValidateAndSetDefaults` (erro específico ou `nil`).

| Caso | Classe | Esperado |
| --- | --- | --- |
| CP01 | todos preenchidos | válido |
| CP02 | 1 caractere em cada campo + 1 scope (limites) | válido |
| CP03 | vários scopes | válido |
| CP04 | tudo ausente | inválido |
| CP05–CP07 | falta só token-url / client-id / client-secret | inválido |
| CP08 | lista de scopes ausente (`nil`) | inválido |
| CP09 | lista de scopes vazia (`[]string{}`) | inválido |
| CP10 | apenas scopes preenchido | inválido |
| CP11 | lista vazia com capacidade (`make([]string,0,2)`) | inválido |
| CP12 | campos só com espaço | **aceito** (observação) |
| CP13 | scopes `[""]` | **aceito** (observação) |
| CP14 | token-url que não é URL (`not a url`) | **aceito** (observação) |

**Lacunas encontradas pela caixa-preta** que a caixa-branca não motivava: (a) a distinção `nil` × lista vazia × vazia com capacidade — para `len` são iguais, mas no YAML correspondem a `scopes` ausente e `scopes: []`; (b) CP12–CP14 mostram que a regra verifica **presença**, não conteúdo: espaço, scope vazio e URL sem formato passam. Isso não contradiz a especificação ("must define all fields"), por isso não foi tratado como falha; os testes registram o comportamento atual e o ponto fica para decisão do grupo (possível ciclo TDD extra).

Resultado conjunto: **4 funções, 22 subtestes PASS**, exit 0.

## Cobertura antes/depois

Seleção da equipe `^TestConfig_` em `./client`, `-covermode=atomic`.

| Momento | `isValid` (OAuth2) | `ValidateAndSetDefaults` | Pacote `client` | Relatório |
| --- | --- | --- | --- | --- |
| Antes (11:28:05) | 0,0% | 69,2% | 12,3% | [antes](evidencias/antes/client-config.cover-func.txt) |
| Só CB, seleção F5 (11:29:40) | 100,0% | 61,5% | 3,0% | [cb](evidencias/cb/client-config.cover-func.txt) |
| CB + CP, seleção da equipe (11:30:21) | 100,0% | 76,9% | 12,7% | [cb-cp](evidencias/cb-cp/client-config.cover-func.txt) |

As execuções da seleção `^TestConfig_` terminam com exit 1 apenas por **CP11 da F1** (`://8.8.8.8:53`), falha conhecida da parte 1 de surpesaiajin, corrigida no TDD dele (ainda não integrado na `ptoss2`). Os testes F5 passam em todas as execuções. Saídas completas: [execucoes.txt](evidencias/execucoes.txt).

## Comparação das técnicas

- A **caixa-branca** já levou `isValid` a 100% de statements com 5 casos e provou MC/DC das quatro condições — mas a cobertura de statements de uma função de uma linha é 100% com um único caso; só a análise de decisão mostra que são necessários os cinco.
- A **caixa-preta** não acrescentou statements (continua 100%), mas trouxe classes de entrada que a leitura do código tratava como iguais (`nil`/`[]`/`[""]`, espaço em branco) e explicitou o limite do que a regra garante. Também exercitou o caminho público e o erro específico.
- Conclusão: cobertura automática não diferencia esses casos; as duas perspectivas se complementam — uma garante as decisões, a outra o significado das entradas.

## Como executar

```sh
go test -mod=readonly -count=1 -timeout=2m -v -run '^TestConfig_PTOSS2_F5_' ./client
go test -mod=readonly -count=1 -timeout=2m -covermode=atomic -coverprofile=client.cover.out -run '^TestConfig_' ./client
go tool cover -func=client.cover.out
```

## Apoio de IA

Testes, medições e este texto foram produzidos com o Claude Code (Anthropic, modelo Claude Opus 5.5) a partir das orientações da equipe, sob revisão de gccintra.
