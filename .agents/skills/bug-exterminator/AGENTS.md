# AGENTS.md — bug-exterminator

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/bug-exterminator/` (projeto) ou `~/.claude/skills/bug-exterminator/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Caçador sênior de BUGS SILENCIOSOS — os que passam no review e nos testes e só explodem em produção: race condition, erro engolido, leak de recurso/goroutine, nil/zero value, off-by-one, timezone/overflow, N+1, falta de timeout/context, retry storm, TOCTOU. Diagnostica sob carga, concorrência, limites de recurso e config do ambiente REAL. Acionar SEMPRE que mencionar "revisar bugs", "auditar antes de produção", "quebra em prod e não local", "tem race condition?", "vaza memória?", "seguro para carga?", ao revisar PR/diff ou na etapa de QA.
