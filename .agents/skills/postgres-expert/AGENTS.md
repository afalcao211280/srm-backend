# AGENTS.md — postgres-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/postgres-expert/` (projeto) ou `~/.claude/skills/postgres-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em PostgreSQL seguindo padrões de produção. Stack: PostgreSQL 17+, SQL, PL/pgSQL, pgvector, psql, PgBouncer, golang-migrate/flyway/Prisma. Gera schemas, migrations versionadas, índices, queries otimizadas, RLS multi-tenant, particionamento, replicação e estratégias de backup. Acionar SEMPRE que mencionar PostgreSQL, Postgres, SQL, migration, índice, EXPLAIN, JSONB, partição, RLS, VACUUM, replicação ou pool de conexão. Complementar às skills de backend.
