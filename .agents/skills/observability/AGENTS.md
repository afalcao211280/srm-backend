# AGENTS.md — observability

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/observability/` (projeto) ou `~/.claude/skills/observability/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Observabilidade: os três pilares (logs, métricas, traces) com OpenTelemetry, Prometheus e logging estruturado. Use SEMPRE ao instrumentar serviços, adicionar métricas/tracing, propagar correlation/trace IDs, configurar dashboards ou alertas. Acionar em "observabilidade", "tracing", "métricas", "Prometheus", "OpenTelemetry", "correlation id", "Grafana", "alerta", "health check".
