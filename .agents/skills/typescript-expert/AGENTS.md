# AGENTS.md — typescript-expert

Skill **portável e self-contained**, pronta para múltiplos agentes de IA. Fonte única: `SKILL.md` (entrada) + `reference/*.md` (sob demanda). Qualquer agente que leia Markdown consome direto — **sem ferramenta de geração**.

## Uso por agente

| Agente | Como usar |
|---|---|
| Claude Code | Copiar a pasta para `.claude/skills/typescript-expert/` (projeto) ou `~/.claude/skills/typescript-expert/` (global). Aciona pela `description` do frontmatter. |
| Cursor | Referenciar `SKILL.md` em `.cursor/rules/*.mdc` ou @-mencionar no chat. |
| GitHub Copilot | Colar o conteúdo de `SKILL.md` em `.github/copilot-instructions.md`. |
| Windsurf / Cline | Copiar `SKILL.md` para o arquivo de regras do agente. |
| Opencode / genérico | Manter `AGENTS.md` + `SKILL.md` na raiz; apontar o agente para a pasta. |

## Regra de ouro

`SKILL.md` é a fonte única de verdade. `reference/` carrega sob demanda (progressive disclosure). Nada aqui depende de binário externo.

## Escopo

Especialista em TypeScript seguindo padrões de produção (cross-cutting, ADR-008). Stack: TypeScript 5.x strict, Zod 4, branded types, discriminated unions, type guards, ts-pattern, type-fest, tsd/expect-type, ESLint `@typescript-eslint`. Complementar às skills de framework (react/nextjs/vue/ angular). Gera tipos, schemas Zod, type guards e configs prontos pra produção. Acionar SEMPRE que mencionar TypeScript,.ts/.tsx, tsconfig, tipagem, generics, interface, type, branded type, Zod, type guard, discriminated union.
