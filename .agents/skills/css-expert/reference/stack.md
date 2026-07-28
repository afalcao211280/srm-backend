# CSS Expert — Stack Canônica

## Tailwind CSS 3.x + shadcn/ui — Versão Canônica Frontend

| Categoria | Lib | Versão | Link | Notas |
|-----------|-----|--------|------|-------|
| **Core CSS** | | | | |
| Utility Framework | Tailwind CSS | 3.x | https://tailwindcss.com | Padrão. v4 em avaliação |
| CSS Variables | CSS Custom Properties | built-in | — | Base para temas shadcn/ui |
| Componentes | shadcn/ui | latest | https://ui.shadcn.com | Copia código. Radix UI + Tailwind |
| Primitivos | Radix UI | 1.x | https://www.radix-ui.com | Headless, acessível, sem estilo |
| **Merge / Variants** | | | | |
| Class merge | clsx + tailwind-merge | 2.x + 2.x | https://github.com/lukeed/clsx | `cn()` helper obrigatório |
| Class variants | class-variance-authority | 0.7.x | https://cva.style | `cva()` para variantes de componentes |
| **Animações** | | | | |
| Animação | Framer Motion | 11.x | https://www.framer.com/motion | Quando necessário. Client only |
| CSS Animations | tailwindcss-animate | 1.x | https://github.com/jamiebuilds/tailwindcss-animate | `animate-*` classes básicas |
| **Ícones** | | | | |
| Ícones | Lucide React | latest | https://lucide.dev | Padrão shadcn/ui |
| Alternativa | Heroicons | 2.x | https://heroicons.com | Heroicons v2 (MIT) |
| **Fontes** | | | | |
| Fontes | next/font (Next.js) | built-in | — | `Inter` padrão. Zero layout shift |
| Fontes genéricas | @fontsource | latest | https://fontsource.org | Para projetos não-Next |
| **Responsividade** | | | | |
| Breakpoints | Tailwind defaults | built-in | — | `sm:640 md:768 lg:1024 xl:1280 2xl:1536` |
| Container queries | `@tailwindcss/container-queries` | 0.1.x | — | Quando necessário |
| **Qualidade** | | | | |
| Class ordering | `prettier-plugin-tailwindcss` | 0.6.x | https://github.com/tailwindlabs/prettier-plugin-tailwindcss | Ordenação automática de classes |

## tailwind.config.ts — Template

```typescript
import type { Config } from 'tailwindcss';
import { fontFamily } from 'tailwindcss/defaultTheme';

const config: Config = {
darkMode: ['class'],
content: [
'./src/**/*.{ts,tsx,js,jsx}',
'./app/**/*.{ts,tsx}',
'./components/**/*.{ts,tsx}',
],
theme: {
extend: {
fontFamily: {
sans: ['var(--font-sans)',...fontFamily.sans],
},
colors: {
border: 'hsl(var(--border))',
input: 'hsl(var(--input))',
ring: 'hsl(var(--ring))',
background: 'hsl(var(--background))',
foreground: 'hsl(var(--foreground))',
primary: {
DEFAULT: 'hsl(var(--primary))',
foreground: 'hsl(var(--primary-foreground))',
},
secondary: {
DEFAULT: 'hsl(var(--secondary))',
foreground: 'hsl(var(--secondary-foreground))',
},
destructive: {
DEFAULT: 'hsl(var(--destructive))',
foreground: 'hsl(var(--destructive-foreground))',
},
muted: {
DEFAULT: 'hsl(var(--muted))',
foreground: 'hsl(var(--muted-foreground))',
},
accent: {
DEFAULT: 'hsl(var(--accent))',
foreground: 'hsl(var(--accent-foreground))',
},
},
borderRadius: {
lg: 'var(--radius)',
md: 'calc(var(--radius) - 2px)',
sm: 'calc(var(--radius) - 4px)',
},
},
},
plugins: [require('tailwindcss-animate')],
};

export default config;
```

## globals.css — CSS Variables shadcn/ui (base)

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
:root {
--background: 0 0% 100%;
--foreground: 222.2 84% 4.9%;
--card: 0 0% 100%;
--card-foreground: 222.2 84% 4.9%;
--popover: 0 0% 100%;
--popover-foreground: 222.2 84% 4.9%;
--primary: 222.2 47.4% 11.2%;
--primary-foreground: 210 40% 98%;
--secondary: 210 40% 96.1%;
--secondary-foreground: 222.2 47.4% 11.2%;
--muted: 210 40% 96.1%;
--muted-foreground: 215.4 16.3% 46.9%;
--accent: 210 40% 96.1%;
--accent-foreground: 222.2 47.4% 11.2%;
--destructive: 0 84.2% 60.2%;
--destructive-foreground: 210 40% 98%;
--border: 214.3 31.8% 91.4%;
--input: 214.3 31.8% 91.4%;
--ring: 222.2 84% 4.9%;
--radius: 0.5rem;
}

.dark {
--background: 222.2 84% 4.9%;
--foreground: 210 40% 98%;
/*... dark mode tokens... */
}
}

@layer base {
* {
@apply border-border;
}
body {
@apply bg-background text-foreground;
}
}
```
