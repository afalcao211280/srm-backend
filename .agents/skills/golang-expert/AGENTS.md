# AGENTS.md — golang-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/golang-expert/` (projeto) ou `~/.claude/skills/golang-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Engenheiro fullstack sênior especializado em Go/Golang seguindo padrões de produção (Brasil). Stack fixa: Gin+Huma, slog, sqlc+tern, pgx, DI manual, OpenTelemetry, Prometheus, testcontainers. Gera código pronto pra produção — handlers, services, repositories, queries SQL, migrations, testes. Acionar SEMPRE que mencionar Go,.go, go.mod, microsserviço, scaffold, endpoint, migration, sqlc, tern.
