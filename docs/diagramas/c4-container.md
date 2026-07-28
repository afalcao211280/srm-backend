# C4 Nível 2 — Container

```
[Browser do Operador]
    |
    | HTTPS (origem do frontend)
    v
[Frontend Next.js 16]
    |
    | reescrita /api/* -> ${API_INTERNAL_URL}/api/* (proxy mesma origem)
    v
[Backend Go (Gin)]
    |
    |-- pgx ---> [PostgreSQL]
    |
    |-- HTTP ----> [Provedor de Câmbio Externo]
    |   (com gobreaker e timeout)
```

Containers:
- **Frontend Next.js 16**: SPA com App Router. Serve o painel e o grid. Reescreve `/api/*` para o backend.
- **Backend Go (Gin)**: API REST em 3 camadas (handler → service → repository → domain). Camada `report` em 2 camadas. Inclui middleware de correlação, exception handler, métricas Prometheus.
- **PostgreSQL**: schema com 6 tabelas (moedas, tipos_recebivel, cedentes, cotacoes_cambio, taxas_base, transacoes + transacao_auditoria). Precisão `NUMERIC` em todos os valores monetários.
- **Provedor de Câmbio Externo**: mockado, atrás de interface substituível. Protegido por circuit breaker.

Protocolos:
- Browser ↔ Frontend: HTTPS, mesma origem.
- Frontend ↔ Backend: HTTP, via nome de serviço do Compose (rede interna).
- Backend ↔ PostgreSQL: pgx (TCP).
- Backend ↔ Provedor: HTTP com timeout explícito.
