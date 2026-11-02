# PTOSS-2 - I2: TDD da configuracao dos badges

**Responsável:** Bercacos. **Branch:** `bercacos-ptoss2`.
**Base:** `14c3db446f45b3af91088a8f203be87d34003986`.
**Execução:** 02/11/2026, Go 1.27.3 windows/amd64, UTC-03.

A [F2](../f2/README.md) registra a parte 1. Os testes novos estão em [TDD1](../../../config/endpoint/ui/ui_badge_ptoss2_tdd1_test.go) e [TDD2](../../../config/endpoint/ui/ui_badge_ptoss2_tdd2_test.go); a implementação está em [ui.go](../../../config/endpoint/ui/ui.go).

## Limitação do histórico

Os testes foram escritos antes das alterações, mas a indisponibilidade inicial do Go impediu executar o Red durante o desenvolvimento. A execução de 12:52:43 reproduziu os estados preservados posteriormente. Ela demonstra as falhas esperadas e a eficácia das correções; não é evidência histórica de execução Red antes da implementação. A integração prévia das cinco partes 1 não foi confirmada. Não houve commit, push ou merge do agente.

## Resultados reproduzidos

| Etapa | PASS / FAIL (subtestes) | Exit | Resultado |
| --- | --- | --- | --- |
| TDD1 Red | 2 / 1 | 1 | Panic de ponteiro nulo no caso de ResponseTime ausente |
| TDD1 Green | 20 / 0 | 0 | Guarda de ResponseTime; 100% de statements |
| TDD2 Red | 6 / 3 | 1 | Três listas negativas aceitas com erro nil |
| TDD2 Green | 29 / 0 | 0 | Negativos rejeitados; 100% de statements |

[Execuções e regressão](evidencias/execucoes.txt). [Cobertura final](evidencias/ui.cover-func.txt). Os percentuais de cobertura parciais nos Reds usam seleções menores de testes e não representam queda da cobertura da suíte completa.

## Parte 2 — mudanças e reprodução dos dois ciclos

### Ciclo 1: ResponseTime ausente

Teste em `ui_badge_ptoss2_tdd1_test.go`: `Badge{}` deve retornar `ErrInvalidBadgeResponseTimeConfig`, sem panic e sem modificar a configuração. Controles: badge ausente ainda recebe padrões e ResponseTime presente com slice nil continua inválido.

Red reproduzido sobre `estados/00-original.go.txt`: o acesso ao ponteiro nulo provocou panic, capturado pelo teste e convertido em falha explícita. Os outros dois controles passaram. Green reproduzido: a verificação `ResponseTime == nil` antes do acesso aos limiares, com curto-circuito na mesma decisão de quantidade, eliminou a falha; todos os 20 subtestes desse estágio passaram. Estado preservado em `estados/01-tdd1.go.txt`.

MC/DC da decisão nova `A || B`: A = ResponseTime nulo; B = quantidade diferente de cinco. Pares: nulo versus cinco valores (A muda, B não é avaliável quando A é verdadeiro: par de masking MC/DC), e ResponseTime presente vazio versus cinco valores (A falso, B muda). Não alegamos unique-cause MC/DC para acesso indefinido com ponteiro nulo.

Refactor: nenhuma alteração adicional necessária no produto; a guarda mínima usa o erro existente. O helper compartilhado evita duplicar as verificações de configuração. O estado Green foi aprovado na reprodução; não houve um terceiro estado de produto para refatoração.

### Ciclo 2: limiares negativos

Teste em `ui_badge_ptoss2_tdd2_test.go`: rejeitar `[-1,0,1,2,3]`, `[-5,-4,-3,-2,-1]` e `[-1,-1,-1,-1,-1]`. Preservar zero, igualdade positiva, quantidade e ordem. Red reproduzido sobre `01-tdd1.go.txt`: os três casos negativos retornaram nil indevidamente e falharam; os seis controles passaram.

Green reproduzido: depois da quantidade, rejeitar primeiro limiar menor que zero. Justificativa: se a lista está ordenada, o primeiro valor é seu mínimo; se um negativo aparece depois de um primeiro valor não negativo, existe necessariamente uma inversão e a regra de ordem já rejeita a lista. Portanto, essa guarda com a validação de ordem cobre todas as posições negativas. Zero e igualdade continuam válidos. Todos os 29 subtestes finais passaram, com 100% de statements nas duas funções do pacote isolado.

Refactor: manter o laço de ordem existente e uma única guarda adicional; não há motivo para reestruturar o método curto. O comentário documenta a relação com o mínimo. A mensagem de erro existente é preservada por compatibilidade. A regressão do pacote UI passou; a regressão mais ampla é registrada separadamente abaixo.

## Como concluir a validação

Com Go compatível com `go.mod` instalado, na raiz do repositório:

```powershell
# Estados isolados: herdados, CP, CP+CB, Red/Green de cada mudança.
./docs/ptoss2/tdd-i2/reproduzir.ps1
# Se necessário: -GoExecutable 'C:\caminho\go.exe'

gofmt -w config/endpoint/ui/ui_badge_ptoss2_*test.go
go test ./config/endpoint/ui -count=1 -v -coverprofile=badge-coverage.out
go tool cover -func=badge-coverage.out
go test ./config/endpoint/... -count=1
git diff --check
```

O script produz data, revisão Git, hashes dos arquivos (as mudanças locais não estão em commit), comandos, saídas, códigos de retorno e cobertura por etapa. Verifica a mensagem esperada do Red, não apenas um exit diferente de zero. Executa uma cópia isolada do pacote UI sem dependências externas; a execução na árvore real acima complementa a reprodução. A suíte ampla pode exigir dependências e serviços adicionais. A comparação quantitativa acima registra a execução já concluída; novas reproduções geram diretórios de evidências separados, contendo apenas o log consolidado e resumos de cobertura; as cópias de código são geradas no diretório temporário do Windows.

## Validação adicional do agente em 02/11/2026

Após a reprodução executada por Bercacos, o agente aplicou `gofmt` aos quatro arquivos novos de testes e repetiu `go test ./config/endpoint/ui -count=1 -v` com cobertura na árvore real. **Os 29 subtestes passaram, com 100% de statements**. A saída está em `evidencias/execucoes.txt`, o resumo por função em `evidencias/ui.cover-func.txt`. Os hashes da reprodução anterior identificam os arquivos daquela execução; a formatação posterior altera seus bytes sem mudar os casos.

Na coleta inicial da cobertura, os testes passaram, mas houve erro de permissão ao renomear um arquivo temporário. A repetição com `GOTMPDIR` em diretório gravável concluiu a coleta. Não se interpreta falha de infraestrutura como falha de comportamento.

O comando `go test ./config/endpoint/... -count=1` **não concluiu a regressão ampla**. A primeira tentativa não pôde gravar o cache de módulos padrão (`evidencias/execucoes.txt`). A repetição com `GOMODCACHE` gravável chegou ao download, mas a resolução DNS de `proxy.golang.org` falhou (`evidencias/execucoes.txt`). Os pacotes UI e SSH passaram; heartbeat não tem testes; endpoint e DNS ficaram com falha de preparação, sem executar seus testes. Repetir o comando em terminal com acesso às dependências antes de afirmar aprovação da regressão ampla.

## Texto-base para o relatório

Minha contribuição cobre a validação dos limiares dos badges de tempo de resposta. Usei primeiro classes de equivalência e limites de quantidade, incluindo ausência do badge, quatro/cinco/seis limiares, ordem, zero e igualdade. Complementei com análise de decisões, listas vazias e inversões em cada posição adjacente. A caixa-preta define o contrato e a caixa-branca identifica caminhos que os casos iniciais não distinguem. Os testes verificam também o estado resultante, pois observar apenas o erro não comprova aplicação dos padrões nem preservação da configuração.

Implementei duas mudanças: devolver erro para ResponseTime ausente e rejeitar negativos preservando zero e igualdade. Após instalar Go 1.27.3 no Windows, executei os testes e a reprodução dos estados preservados. O primeiro Red detectou panic por ponteiro nulo; o segundo detectou aceitação indevida de três listas negativas. As respectivas correções fizeram os testes passarem. A suíte final contém 29 subtestes e obteve 100% de statements no pacote isolado de UI. A cobertura já era 100% nos testes herdados; portanto, o ganho foi de diversidade de casos e força dos oráculos, sem aumento percentual.

A dificuldade inicial de ambiente impediu executar o Red durante a implementação. Os registros atuais são uma reprodução posterior e não devem ser apresentados como histórico completo de TDD. A coordenação de integração da parte 1 e a avaliação dessa limitação pelo grupo permanecem pendentes.

**Apoio de IA:** um agente Codex auxiliou na leitura das orientações e do código, seleção e escrita dos testes, alterações do produto, preparação do script e documentação. Bercacos instalou o Go, executou os testes e o script e compartilhou as saídas para conferência. O agente conferiu os resultados e atualizou este registro. A revisão de Bercacos e a discussão dos resultados com o grupo permanecem necessárias. O texto distingue a execução realizada da reprodução posterior dos estados.
