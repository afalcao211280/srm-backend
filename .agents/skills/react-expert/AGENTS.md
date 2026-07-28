# AGENTS.md — react-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/react-expert/` (projeto) ou `~/.claude/skills/react-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

`requires`: security-expert, typescript-expert — instale junto.

## Escopo

"Engenheiro frontend sênior especializado em React seguindo padrões de produção. Use SEMPRE que houver menção a React, arquivos .tsx/.jsx, componentes React, hooks, ou quando o contexto for um projeto React. Stack: React 18+, TypeScript, React Query/TanStack Query, Zustand ou Redux Toolkit, React Router, Vitest + RTL, Storybook, Tailwind CSS. Arquitetura: feature-based folders, componentes compositions, hooks customizados."
