# ADR-016 — Proxy de mesma origem no frontend

**Decisão:** `next.config.ts` reescreve `/api/:path*` para `${API_INTERNAL_URL}/api/:path*`. O browser fala apenas com a origem do frontend. O Next encaminha ao backend pela rede interna do Compose.

**Consequência:** sem requisição cross-origin partindo do browser, CORS deixa de ser problema a configurar. A porta do backend continua publicada no Compose para acesso direto a Swagger, métricas e health checks durante a avaliação — conveniência de ambiente, não parte do caminho funcional.
