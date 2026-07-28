---
name: caching-strategy
description: >
Estratégias de cache com Redis: cache-aside, TTL, invalidação, multi-nível e
HTTP caching. Use SEMPRE ao otimizar performance, reduzir carga de banco, cachear
respostas, ou lidar com invalidação e cache stampede. Acionar em "cache", "Redis",
"cache-aside", "invalidação", "TTL", "performance", "reduzir latência", "memoização".
version: "1.0.0"
category: Crosscutting
keywords:
- caching
- redis
- cache-aside
- invalidation
- ttl
- performance
- http-caching
---

# Caching Strategy (Cross-cutting)

Stack: Redis (go-redis/ioredis/redis-py). Cache é otimização — corretude primeiro.

## Princípios

1. **Cache-aside por padrão** — app lê cache; miss → lê fonte → popula cache. Simples e resiliente.
2. **TTL sempre** — todo item expira. Sem TTL = memória infinita + dados stale eternos.
3. **Invalidação explícita na escrita** — write → invalida (ou atualiza) a chave. Prefira invalidar a manter coerência complexa.
4. **Chaves versionadas/namespaced** — `svc:entity:id:v1`. Facilita invalidação em massa por prefixo.
5. **Degradação graciosa** — cache down ≠ app down. Fallback pra fonte.

## Padrões

| Padrão | Quando |
|---|---|
| Cache-aside (lazy) | Leitura dominante, tolerância a stale curto |
| Write-through | Consistência forte no cache, escrita cara aceitável |
| Write-behind | Alta escrita, tolera perda em crash (cuidado) |
| Multi-nível (L1 local + L2 Redis) | Latência ultrabaixa, hot keys |
| HTTP caching (`Cache-Control`, `ETag`) | Respostas de API/CDN |

## Invalidação

- **TTL**: expiração natural (escolher por volatilidade do dado).
- **Event-based**: invalida na mutação (evento de domínio → delete key).
- **Versão/tag**: bump de versão no prefixo invalida grupo.

## Cache Stampede (thundering herd)

- Miss simultâneo em chave popular → sobrecarrega a fonte.
- Mitigar: lock distribuído (`SET NX`) p/ um recomputar; TTL jitter; early-recompute; `singleflight` (Go).

## Anti-padrões

- ❌ Cache sem TTL
- ❌ Cachear dados sensíveis sem cuidado (PII em cache compartilhado)
- ❌ Invalidação por "adivinhação" (chaves inconsistentes)
- ❌ App quebra se Redis cai (sem fallback)
- ❌ Chave sem namespace (colisão)
- ❌ Cachear resposta de erro

## Checklist

- [ ] TTL definido por chave
- [ ] Invalidação na escrita
- [ ] Chaves namespaced/versionadas
- [ ] Fallback se cache indisponível
- [ ] Proteção contra stampede em hot keys
- [ ] HTTP `Cache-Control`/`ETag` onde aplicável
- [ ] Nada sensível em cache compartilhado sem controle

## Cross-references

- `golang-expert` (go-redis), `nestjs-expert` (CacheModule), `python-expert`, `java-expert` — implementação
- `postgres-expert` — reduzir carga de leitura
- `observability` — hit ratio, latência de cache

## Seguranca (Baseline Compartilhado)

Cuidado com PII/secrets em cache. Regras universais em `reference/security-baseline.md`.
