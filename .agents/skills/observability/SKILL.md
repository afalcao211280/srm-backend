---
name: observability
description: >
Observabilidade: os três pilares (logs, métricas, traces) com OpenTelemetry,
Prometheus e logging estruturado. Use SEMPRE ao instrumentar serviços, adicionar
métricas/tracing, propagar correlation/trace IDs, configurar dashboards ou alertas.
Acionar em "observabilidade", "tracing", "métricas", "Prometheus", "OpenTelemetry",
"correlation id", "Grafana", "alerta", "health check".
version: "1.0.0"
category: Crosscutting
keywords:
- observability
- opentelemetry
- prometheus
- tracing
- metrics
- logging
- grafana
---

# Observability — Três Pilares (Cross-cutting)

Stack: OpenTelemetry (traces) + Prometheus (métricas) + logging estruturado (slog/Pino). Correlation por trace_id.

## Princípios

1. **Três pilares integrados** — trace_id no log, span no service, métrica no middleware. Correlacionáveis.
2. **Structured logging** — JSON em prod. Nunca `print`/`console.log`. Níveis: debug/info/warn/error.
3. **Trace context propagation** — W3C Trace Context (`traceparent`). Propagar em toda chamada outbound (HTTP/fila).
4. **Correlation ID** — gerado no ingress (middleware), carregado no `context`, presente em todo log/span.
5. **RED/USE** — métricas de serviço: Rate, Errors, Duration. Recursos: Utilization, Saturation, Errors.

## Stack Canônica

| Pilar | Ferramenta | Uso |
|---|---|---|
| Traces | OpenTelemetry (OTLP gRPC) | Spans no service, propagação W3C |
| Métricas | Prometheus (`/metrics`) | Counters, histograms, gauges |
| Logs | slog (Go) / Pino (Node) / structlog (Py) | JSON estruturado + trace_id |
| Dashboards | Grafana | RED/USE, SLO |
| Alertas | Prometheus alert rules → Alertmanager/PagerDuty | SLO burn, error rate, latência |
| Health | `/healthz` `/readyz` | liveness/readiness (k8s probes) |

## Padrões

- **Span**: criar no início do use case, `end` no defer; atributos = ids relevantes; status de erro no span.
- **Métricas custom**: histogram de latência por endpoint; counter de erros por tipo; gauge de fila.
- **Log de erro**: incluir `error`, `trace_id`, contexto operacional (nunca PII/secrets).
- **Alertas**: baseados em SLO (ex: p99 > X por 5min; error rate > 1%); evitar alertas ruidosos.

## Anti-padrões

- ❌ Log sem trace_id (não correlaciona)
- ❌ Logar PII/tokens/senhas
- ❌ Métrica de alta cardinalidade (label com id de usuário)
- ❌ Alerta sem SLO (ruído → fadiga)
- ❌ Trace sem propagação outbound (traces quebrados)

## Checklist

- [ ] trace_id no contexto + em todos os logs
- [ ] Span OTel no service, status de erro
- [ ] `/metrics` Prometheus (RED)
- [ ] Propagação W3C em chamadas outbound
- [ ] Health checks liveness/readiness
- [ ] Alertas baseados em SLO
- [ ] Logs JSON sem dados sensíveis

## Cross-references

- `golang-expert`, `nestjs-expert`, `python-expert`, `java-expert` — instrumentação por stack
- `kubernetes-expert` — scrape Prometheus, probes
- `security-expert` — logs sem dados sensíveis

## Seguranca (Baseline Compartilhado)

Nunca logar PII/secrets. Regras universais em `reference/security-baseline.md`.
