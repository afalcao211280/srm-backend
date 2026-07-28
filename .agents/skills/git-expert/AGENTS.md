# AGENTS.md — git-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/git-expert/` (projeto) ou `~/.claude/skills/git-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em Git workflow, versionamento e automação. Use SEMPRE que houver commits, branches, pull requests/PR, merge, rebase, tags, releases, changelog, SemVer, hooks, gitflow, "criar/revisar PR", "gerar release", "validar commits", "bump version". AGENTIC: gera PR descriptions, checklist de revisão, calcula SemVer, gera changelogs, configura hooks e scripts de release. Complementa ADR-023 (branching) e ADR-028 (hooks).
