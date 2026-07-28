# Critérios de aceite

## Usabilidade

- O operador consegue simular um recebível em menos de 3 cliques após carregar a página.
- O resultado da simulação aparece sem refresh, com debounce de 300ms.
- Erros de validação aparecem ao lado do campo correspondente.
- O grid reflete o estado da URL e é recarregável e compartilhável.

## Segurança

- Toda requisição HTTP carrega um identificador de correlação propagado aos logs e ao corpo de erro.
- Logs não contêm payload completo, credenciais, dados de cedente além do identificador, ou string de conexão.
- Tipos `float32`/`float64` proibidos no domínio e na persistência (verificado por lint).
- `internal/report` não importa `internal/domain` (verificado por `depguard`).
- Toda resposta de erro segue o formato padronizado e nunca vaza stack trace.

## Desempenho

- Cálculo de precificação em <50ms para prazos até 365 dias.
- Latência p95 de listagem do grid <200ms com 10k transações.
- Plano de execução do extrato filtrado usa índice, sem varredura sequencial.

## Escalabilidade

- O design para 1M transações/minuto está documentado em `docs/alta-escala.md`.
- Reconhece os limites da paginação por `LIMIT/OFFSET` e da escrita síncrona.
- Identifica os caminhos de evolução: cache, sharding, keyset pagination, EDA.
