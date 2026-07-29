# SRM Credit Engine — Design de alta escala (1M transações/minuto)

A solução atual não foi dimensionada para 1M transações/minuto. Este documento registra o que precisaria mudar.

## Gargalos identificados

- **Paginação por deslocamento** em `LIMIT/OFFSET`: degrada em profundidade. Acima de ~100k linhas por filtro, o custo do `OFFSET` cresce linearmente com a posição.
- **Escrita síncrona na liquidação**: cada liquidação faz um `UPDATE` em banco relacional com optimistic locking. A contenção na linha individual é baixa, mas o total agregado de I/O no banco pode ser o gargalo.
- **Pool de conexões pgx**: precisa crescer com a concorrência, mas o banco é o limite físico.

## Caminhos de evolução

- **Cache de leitura** com TTL curto para a listagem do grid, com invalidação por evento.
- **Keyset pagination** com cursor opaco no lugar de `OFFSET`.
- **Sharding** por cedente ou por mês de operação.
- **Consistência eventual**: o extrato analítico não precisa de consistência forte; CDC + read replica é viável.
- **EDA** (proposta no `eda.md`): liquidação publica evento; integrações reagem assincronamente.

## Reconhecimento dos limites

A paginação por deslocamento e a escrita síncrona são suficientes para a entrega avaliada e inadequadas para a escala-alvo. A escolha é deliberada: a otimização prematura custa complexidade sem retorno em volume de produção atual.
