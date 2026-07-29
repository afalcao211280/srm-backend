# Proposta de arquitetura orientada a eventos (EDA)

Esta é uma proposta documental. Nenhum broker precisa ser implantado.

## Eventos de domínio

| Evento | Produtor | Consumidores |
|---|---|---|
| `TransacaoCriada` | `POST /api/v1/transacoes` | Auditoria, BI, notificações |
| `TransacaoLiquidada` | `POST /api/v1/transacoes/{id}/liquidacao` | Conciliação contábil, BI |
| `TransacaoCancelada` | cancelamento manual | BI |
| `CotacaoRegistrada` | `POST /api/v1/cotacoes` | Cache de cotação vigente |
| `TaxaBaseRegistrada` | `POST /api/v1/taxas-base` | Cache de taxa vigente |

## Garantias de entrega

- **Idempotência**: cada evento carrega `id` da transação. Consumidores deduplicam por `id`.
- **Ordenação**: eventos da mesma transação são ordenados por `versao` e `instante`. Consumidores aplicam na ordem.
- **Outbox**: a publicação de evento acontece na mesma transação de banco que registra a alteração — sem 2PC.
