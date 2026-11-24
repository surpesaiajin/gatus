# EX04 — Observações no painel

Observação em 23/11/2026, commit `cdcd5721`, porta local 18080. As capturas foram geradas com o Microsoft Edge em modo headless (`--screenshot`, janela de 1280 × 1100) apontado para `http://127.0.0.1:18080/`, com **Filter by → None** e busca vazia.

| Hora local | Estado do arquivo | Painel | Evidência |
| --- | --- | --- | --- |
| 21:36:56 | Original (`recarga-original`, `interval: 3s`) | Seis cartões. `recarga-original` (grupo `configuracao`) em `Healthy` | [01-painel-inicial.png](01-painel-inicial.png) |
| 21:38:42 | `recarga-nova`, `interval: 3s` | Seis cartões. O primeiro passou a ser `recarga-nova` (grupo `configuracao`), em `Healthy`. Nenhum cartão `recarga-original` | [02-painel-apos-recarga-valida.png](02-painel-apos-recarga-valida.png) |
| 21:40:35 | `recarga-nova`, `interval: invalido` | Sem página. O Edge headless terminou sem gerar o arquivo de captura, e a navegação pelo navegador embutido do ambiente de trabalho, por volta de 21:41, também falhou ao abrir o endereço. A indisponibilidade está comprovada pelo `curl` | [curl-cenario2-indisponivel.txt](curl-cenario2-indisponivel.txt) |
| 21:44:31 | Original restaurado e data de modificação atualizada | Seis cartões. `recarga-original` de volta, em `Healthy`, com a barra de histórico reiniciada (poucos checks verdes) | [04-painel-apos-restauracao.png](04-painel-apos-restauracao.png) |

O nome do serviço ativo é identificável pelo título do cartão antes e depois de cada recarga. O painel não mostra aviso de que houve recarga nem de que uma recarga falhou. Durante a falha, não houve página alguma para mostrar esse aviso.
