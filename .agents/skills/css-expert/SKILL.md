---
name: css-expert
description: >
Especialista em CSS e design systems seguindo padrões de produção
(cross-cutting, ADR-008). Stack: Tailwind CSS 3.4+, CSS Custom Properties
(design tokens), CSS Modules, PostCSS, shadcn/ui + Radix UI, cva + cn()
(clsx+tailwind-merge), Lucide/Heroicons. Gera componentes estilizados,
tokens, layouts responsivos, dark mode e a11y prontos pra produção. Acionar
SEMPRE que mencionar CSS, Tailwind, estilo, responsivo, layout, dark mode,
design system, design token, animação CSS, shadcn, ou estilização de UI.
version: "2.1.0"
category: Frontend
keywords:
- css
- tailwind
- design-tokens
- responsive
- dark-mode
- shadcn
- accessibility
- postcss
- design-system
requires:
- security-expert
---

# CSS Expert — Padrões

Especialista em CSS e design systems. Utility-first com Tailwind, design tokens como fundação, zero estilo mágico. Código que entra em produção.

## Princípios

1. **Utility-first** — composição de classes Tailwind. CSS custom só quando utilitário não cobre. Elimina CSS morto e inconsistência de naming.
2. **Design tokens como fundação** — CSS Custom Properties em `:root`; Tailwind `theme.extend` referencia os tokens. Nunca valor hardcoded em componente — tokens permitem theming (dark mode, white-label).
3. **Mobile-first** — base é mobile; `sm:`/`md:`/`lg:`/`xl:` para telas maiores. Nunca desktop-first.
4. **Dark mode via `class`** — `darkMode: 'class'`; tokens semânticos respondem ao `.dark`. Preferir tokens semânticos (`bg-background`) a cores fixas com variant `dark:`.
5. **Acessibilidade não é opcional** — contraste WCAG AA (4.5:1 texto), `focus-visible`, `prefers-reduced-motion`, `sr-only`. Componentes interativos complexos via Radix (a11y nativa).
6. **`cn()` sempre** — `clsx` + `tailwind-merge` resolvem conflito de classes (`p-2 p-4` → `p-4`). Concatenação crua de string produz bugs de ordem.

## Stack Canônica

| Camada | Lib | Uso |
|---|---|---|
| Utility CSS | Tailwind CSS 3.4+ | utility-first, JIT, mobile-first |
| Design tokens | CSS Custom Properties | `:root` + `.dark`; base do shadcn/ui |
| Scoped CSS | CSS Modules | zero-runtime, build-time (quando necessário) |
| Processamento | PostCSS | autoprefixer, nesting nativo |
| Componentes | shadcn/ui | copia código (não dependência); Radix + Tailwind |
| Primitivos | Radix UI | headless, acessível, sem estilo |
| Merge de classes | clsx + tailwind-merge | `cn()` helper obrigatório |
| Variantes | class-variance-authority (cva) | variantes tipadas de componente |
| Ícones | Lucide / Heroicons | tree-shakeable, SVG-first |
| Animações | Tailwind + CSS keyframes / tailwindcss-animate | cross-framework, CSS-native |
| Fontes | next/font / @fontsource | zero layout shift |
| Class ordering | prettier-plugin-tailwindcss | ordenação automática |

> **FIXO** (ADR-008): Tailwind utility-first, CSS Custom Properties como tokens, mobile-first, dark mode via `class`, PostCSS. `@apply` **apenas** em `src/components/ui/`. **Nota**: Tailwind v4 (Oxide) em avaliação — canônico é 3.4+. **Framer Motion** é React-only; para animação cross-framework use CSS/Tailwind.
> Complementa `react-expert`, `nextjs-expert`, `vue-expert`, `angular-expert`.

## Design Tokens (Obrigatório)

```css
:root {
--color-primary-500: #3b82f6;
--spacing-unit: 0.25rem;
--radius-md: 0.375rem;
--font-size-base: clamp(1rem, 1rem + 0.5vw, 1.125rem);
}
.dark { --color-primary-500: #60a5fa; }
```

Tokens em CSS Custom Properties; Tailwind referencia via `theme.extend`. Ver `reference/stack.md` para o setup shadcn/ui completo (`tailwind.config.ts` + `globals.css`).

## Padrões Obrigatórios

- **Responsive mobile-first**: `sm:`/`md:`/`lg:`/`xl:`. Nunca font-size só por breakpoint — use `clamp()` para fluid typography.
- **Dark mode**: `darkMode: 'class'`, tokens com `.dark` variant. Preferir tokens semânticos a cores fixas.
- **`@apply`** apenas em componentes base (`src/components/ui/`), nunca em features. Para padrão reutilizável, `@layer components`.
- **Layout**: Grid para layouts complexos, Flexbox para alinhamentos simples. `aspect-ratio` para proporções (nunca padding-bottom hack). Container queries (`@container`) para componentes isolados.
- **Variantes de componente** com `cva`; combinar classes com `cn()`. `group`/`peer` para estados de pai/irmão (zero JS).
- **Z-index** em escala nomeada (`z-dropdown`, `z-modal`, `z-toast`) — nunca guerra de `z-index`.
- **a11y**: `focus-visible` (não `:focus` em clique), `prefers-reduced-motion: reduce`, `prefers-color-scheme`, contraste WCAG AA+, `sr-only`, skip-to-content.
- **Performance**: purge por `content` paths; `content-visibility: auto` para offscreen; `will-change` só com profiler; critical CSS inline; budget CSS < 50KB gzip.

## Workflow Agentic

1. **Tokens primeiro** — definir/confirmar design tokens (`:root` + `.dark`) e `tailwind.config` antes de estilizar.
2. **Primitivos** — componentes `ui/` com `cva` + `cn()`; interativos complexos sobre Radix (a11y).
3. **Composição** — features usam utilitários direto no template; nunca `@apply` em feature.
4. **Verificar** — dark mode, breakpoints (mobile→desktop), contraste/foco/reduced-motion, purge em produção, ordenação de classes (Prettier).

## Anti-padrões

| Proibido | Alternativa |
|---|---|
| `!important` / prefixo `!` | especificidade via `cn()` na ordem correta |
| Inline style para valor dinâmico | CSS custom property + classe Tailwind |
| Cor hex arbitrária (`text-[#1a2b3c]`) | token do design system / token semântico |
| Espaçamento mágico (`mt-[13px]`) | escala Tailwind (ou estender o tema) |
| z-index warfare | escala nomeada (`z-dropdown`, `z-modal`) |
| Concatenar classes com template string | `cn()` (clsx + tailwind-merge) |
| Duplicar string longa de classes | `cva` para variantes |
| `@apply` em feature component | utilitários no template / `@layer components` em `ui/` |
| Cores sem variant `dark:` / token semântico | tokens semânticos que respondem ao dark |
| Classe Tailwind dinâmica (`text-${c}-500`) | mapa de classes completas ou `safelist` |
| Layout shift sem fallback | `aspect-ratio`, `min-height`, skeleton |
| Animação sem `prefers-reduced-motion` | `motion-reduce:` variant |
| Hover/estado gerenciado por JS state | `group`/`peer` (CSS puro) |

## Checklist

- [ ] Design tokens em `:root` (+ `.dark`); Tailwind referencia tokens
- [ ] Responsive mobile-first testado (sm→xl)
- [ ] Dark mode funcional (tokens semânticos)
- [ ] `cn()` para classes condicionais; `cva` para variantes
- [ ] a11y: contraste WCAG AA, `focus-visible`, reduced-motion, `sr-only`
- [ ] `@apply` restrito a `src/components/ui/`
- [ ] Purge CSS configurado (zero classe morta em produção)
- [ ] Escala de z-index documentada
- [ ] Budget CSS < 50KB gzip; `prettier-plugin-tailwindcss` no CI

## Quando Perguntar

Antes de estilizar: CSS-in-JS vs Tailwind (contexto do projeto)? UI kit (shadcn/ui vs Vuetify/PrimeVue/Material — conforme framework)? Estratégia de animação (CSS vs Framer Motion, se React)?
Decisões que não se assume: arquitetura do design system (tokens → componentes → padrões), estratégia de browser support (prefixos, fallbacks), migração Tailwind v3 → v4.

## Referências (sob demanda — progressive disclosure)

- **`reference/core.md`** — Padrões de código completos: `cn()` helper, componente com `cva` (Button), layouts (sidebar, grid, stack), dark mode com CSS variables, animações (Framer Motion), a11y com Radix (Dialog), skeleton, ordem de classes Tailwind, + 11 anti-padrões detalhados com o porquê. Leia antes de estilizar.
- **`reference/stack.md`** — Stack canônica: versões, `tailwind.config.ts` template (shadcn/ui), `globals.css` com CSS variables (tokens light/dark). Leia ao configurar projeto novo.

## Cross-references

- `typescript-expert` — tipos de Tailwind config / theme, `VariantProps`
- `react-expert`, `nextjs-expert`, `vue-expert`, `angular-expert` — integração de estilo por framework
- `security-expert` — CSP (defesa contra CSS/style injection)

## Segurança (Baseline Compartilhado)

Regras universais em `reference/security-baseline.md`. Específico de estilização:

- **Nunca `style`/CSS a partir de input de usuário não sanitizado** — CSS injection pode exfiltrar dados (seletores de atributo) ou quebrar layout. Renderizar apenas tokens/classes controladas.
- **CSP** no servidor restringindo `style-src` como defesa em profundidade; evitar `style-src 'unsafe-inline'` quando possível.
- **Ícones/SVG de terceiros**: sanitizar antes de inline (SVG pode conter `<script>`).
