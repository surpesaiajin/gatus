# PTOSS-2 - F2: configuracao dos badges

**Responsável:** Bercacos. **Ordem:** caixa-preta, depois caixa-branca.
**Base:** `14c3db446f45b3af91088a8f203be87d34003986` (`ptoss2`, sobre `d255d0c`).
**Execução:** 02/11/2026, Go 1.27.3 windows/amd64, UTC-03.

Os resultados abaixo foram obtidos na reprodução de 12:52:43 sobre o estado original do método, anterior às correções de UI. Testes mantidos em [caixa-preta](../../../config/endpoint/ui/ui_badge_ptoss2_blackbox_test.go) e [caixa-branca](../../../config/endpoint/ui/ui_badge_ptoss2_whitebox_test.go).

## Parte 1 — caixa-preta seguida de caixa-branca

Contrato: badge ausente recebe `[50,200,300,500,750]`; badge presente exige cinco valores em ordem não decrescente. Zero e igualdade são aceitos. Conferimos erro e configuração resultante: valores, identidade dos objetos existentes e preservação das demais flags de UI. Não há chamadas de rede.

| Caso | Entrada | Expectativa e justificativa |
| --- | --- | --- |
| CP01 | badge ausente | Sucesso, valores padrão; classe de ausência |
| CP02 | `[50,200,300,500,750]` | Sucesso, preservação; classe válida |
| CP03 | 4 valores | Erro; limite inferior de quantidade |
| CP04 | 6 valores | Erro; limite superior de quantidade |
| CP05 | `[50,200,500,300,750]` | Erro; inversão interior |
| CP06 | `[0,1,2,3,4]` | Sucesso; limite zero |
| CP07 | `[0,0,0,0,0]` | Sucesso; igualdade |
| CB01/02 | slice nil / vazio, ResponseTime presente | Erro; comprimentos zero com representações distintas |
| CB03 | `[10,20,30,50,40]` | Erro na comparação i=4 |
| CB04 | `[10,30,20,40,50]` | Erro na comparação i=2 |
| CB05 | `[20,10,30,40,50]` | Erro na comparação i=1 |
| CB06 | `[10,20,20,40,50]` | Sucesso; igualdade dentro de lista não constante |

Primeiro conjunto: `ui_badge_ptoss2_blackbox_test.go`. Complemento: `ui_badge_ptoss2_whitebox_test.go`. Houve leitura prévia do código: a sequência representa a perspectiva de seleção e escrita, não desconhecimento da implementação.

Na versão original, decisões: `Badge != nil` (CP01 falso, CP02 verdadeiro); quantidade diferente de cinco (CP03 verdadeiro, CP02 falso); laço `i > 0` (CP02 entra e termina); comparação adjacente (CP02 falso, CP05 verdadeiro). A caixa-branca acrescenta inversões em todas as posições restantes e slice nil/vazio. A comparação i=3 já é exercitada por CP05.

**MC/DC na parte 1:** as decisões originais têm uma condição cada; seus resultados verdadeiro/falso são suficientes para a condição individual. Não há decisão booleana composta à qual aplicar pares de independência. Cobertura de statements não mede cobertura de decisões nem de iterações. A execução confirmou que o conjunto herdado já percorre todos os statements do método. Mesmo sem aumento percentual, os novos casos fortalecem os oráculos e as posições verificadas.

**Comparação medida em 02/11/2026:** os três primeiros estágios usam o mesmo código original, sem as correções. Os casos CP e CB descritos na tabela passaram.

| Conjunto | Casos (subtestes) | Resultado | ValidateAndSetDefaults | GetDefaultConfig | Total de statements |
| --- | --- | --- | --- | --- | --- |
| Herdados | 4 | PASS | 100% | 100% | 100% |
| Herdados + caixa-preta | 11 | PASS | 100% | 100% | 100% |
| Herdados + caixa-preta + caixa-branca | 17 | PASS | 100% | 100% | 100% |

O aumento foi de **0 ponto percentual**. O benefício foi a ampliação dos comportamentos verificados: os testes herdados, mesmo com 100% de statements, não detectavam ResponseTime ausente nem negativos em ordem crescente. A reprodução dos testes novos contra os estados anteriores demonstrou essas lacunas.

## Evidências e reprodução

[Execuções](evidencias/execucoes.txt). Cobertura por função: [antes](evidencias/antes/ui.cover-func.txt), [CP](evidencias/cp/ui.cover-func.txt) e [CP + CB](evidencias/cp-cb/ui.cover-func.txt).

O script em [TDD de I2](../tdd-i2/reproduzir.ps1) reproduz também os três conjuntos desta F2. Os estados anteriores estão preservados em I2; fontes e testes de trabalho permanecem no pacote UI. A integração das cinco partes 1 antes das mudanças de produto não foi confirmada; a execução dos estados foi posterior à implementação.

## Conclusão e apoio de IA

A cobertura de statements permaneceu em 100%, mas os novos testes ampliaram a verificação do estado resultante, dos limites e das posições de inversão. O [TDD de I2](../tdd-i2/README.md) registra os comportamentos adicionais, resultados e limitações.

Um agente Codex auxiliou no projeto e escrita dos testes, análise de decisões e documentação. Bercacos instalou o Go, executou os testes e compartilhou as saídas para conferência. A revisão e discussão dos resultados com o grupo permanecem necessárias.