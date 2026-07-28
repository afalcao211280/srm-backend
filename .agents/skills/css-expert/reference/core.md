# CSS Expert — Padrões de Código (Tailwind CSS + shadcn/ui)

## 1. `cn()` Helper — Obrigatório

```typescript
// src/lib/utils.ts
import { type ClassValue, clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

// Uso:
// cn('px-4 py-2', condition && 'bg-red-500', { 'opacity-50': disabled })
// twMerge resolve conflitos: cn('px-2 px-4') → 'px-4'
```

---

## 2. Componente com `cva` (Class Variance Authority)

```tsx
// src/components/ui/button.tsx
import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@/lib/utils';

const buttonVariants = cva(
  // Base classes — sempre aplicadas
  'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'border border-input bg-background hover:bg-accent hover:text-accent-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 rounded-md px-3',
        lg: 'h-11 rounded-md px-8',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <button className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
  ),
);
Button.displayName = 'Button';

export { buttonVariants };
```

---

## 3. Layout Patterns com Tailwind

```tsx
// Sidebar + Content layout
export function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      {/* Sidebar — fixed width, full height */}
      <aside className="w-64 shrink-0 border-r bg-sidebar">
        <nav className="flex flex-col gap-1 p-4">
          {/* nav items */}
        </nav>
      </aside>
      {/* Main — takes remaining space */}
      <main className="flex-1 overflow-auto">
        <div className="container mx-auto max-w-7xl px-6 py-8">
          {children}
        </div>
      </main>
    </div>
  );
}

// Card grid responsivo
export function UserGrid({ users }: { users: User[] }) {
  return (
    <ul className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {users.map((user) => (
        <li key={user.id}>
          <UserCard user={user} />
        </li>
      ))}
    </ul>
  );
}

// Stack (vertical) com gap consistente
export function FormStack({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-col gap-4">{children}</div>;
}
```

---

## 4. Dark Mode com CSS Variables

```tsx
// src/components/providers/theme-provider.tsx
'use client';
import { createContext, useContext, useEffect, useState } from 'react';

type Theme = 'dark' | 'light' | 'system';

const ThemeContext = createContext<{
  theme: Theme;
  setTheme: (theme: Theme) => void;
}>({ theme: 'system', setTheme: () => {} });

export function ThemeProvider({ children, defaultTheme = 'system' }: {
  children: React.ReactNode;
  defaultTheme?: Theme;
}) {
  const [theme, setTheme] = useState<Theme>(defaultTheme);

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove('light', 'dark');
    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      root.classList.add(systemTheme);
    } else {
      root.classList.add(theme);
    }
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export const useTheme = () => useContext(ThemeContext);
```

---

## 5. Animações com Framer Motion

```tsx
// src/components/ui/animated-list.tsx
'use client';
import { motion, AnimatePresence } from 'framer-motion';

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: i * 0.05, duration: 0.3, ease: 'easeOut' },
  }),
  exit: { opacity: 0, x: -20, transition: { duration: 0.15 } },
};

export function AnimatedList<T extends { id: string }>({
  items,
  renderItem,
}: {
  items: T[];
  renderItem: (item: T) => React.ReactNode;
}) {
  return (
    <ul className="space-y-2">
      <AnimatePresence initial={false}>
        {items.map((item, i) => (
          <motion.li
            key={item.id}
            custom={i}
            variants={itemVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            layout
          >
            {renderItem(item)}
          </motion.li>
        ))}
      </AnimatePresence>
    </ul>
  );
}
```

---

## 6. Acessibilidade — Padrões

```tsx
// Sempre usar Radix UI para componentes interativos complexos
import * as Dialog from '@radix-ui/react-dialog';
import * as VisuallyHidden from '@radix-ui/react-visually-hidden';

export function ConfirmDialog({ open, onClose, onConfirm, title, description }: ConfirmDialogProps) {
  return (
    <Dialog.Root open={open} onOpenChange={onClose}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 bg-black/50 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <Dialog.Content
          className="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-md rounded-lg bg-background p-6 shadow-lg focus:outline-none"
          aria-describedby="dialog-desc"
        >
          <Dialog.Title className="text-lg font-semibold">{title}</Dialog.Title>
          <Dialog.Description id="dialog-desc" className="mt-2 text-sm text-muted-foreground">
            {description}
          </Dialog.Description>
          <div className="mt-6 flex justify-end gap-3">
            <Dialog.Close asChild>
              <button className="rounded-md px-4 py-2 text-sm hover:bg-accent">Cancelar</button>
            </Dialog.Close>
            <button
              className="rounded-md bg-destructive px-4 py-2 text-sm text-destructive-foreground hover:bg-destructive/90"
              onClick={onConfirm}
            >
              Confirmar
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
```

---

## 7. Skeleton Loading

```tsx
// src/components/ui/skeleton.tsx
import { cn } from '@/lib/utils';

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn('animate-pulse rounded-md bg-muted', className)}
      aria-hidden="true"
    />
  );
}

// Uso:
export function UserCardSkeleton() {
  return (
    <div className="flex items-center gap-4 rounded-lg border p-4">
      <Skeleton className="h-10 w-10 rounded-full" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-48" />
      </div>
      <Skeleton className="h-6 w-16 rounded-full" />
    </div>
  );
}
```

---

## 8. Regras de Organização de Classes Tailwind

```
Ordem recomendada (prettier-plugin-tailwindcss aplica automaticamente):
1. Layout:      block flex grid hidden
2. Position:    relative absolute fixed sticky
3. Display:     inset top right bottom left z-index
4. Flexbox:     flex-col items-center justify-between gap-4
5. Grid:        grid-cols-3 col-span-2
6. Size:        w-full h-10 min-w-0 max-w-xl
7. Margin:      m-4 mx-auto mt-8
8. Padding:     p-4 px-6 py-2
9. Typography:  text-sm font-medium leading-6 tracking-wide
10. Color:      text-primary bg-background border-border
11. Border:     border border-border rounded-lg
12. Effects:    shadow-md opacity-50 ring-2
13. Transition: transition-colors duration-200 ease-in-out
14. Hover/Focus: hover:bg-accent focus-visible:ring-2
15. Responsive:  sm:flex-row md:grid-cols-2 lg:hidden
16. Dark:        dark:bg-gray-800 dark:text-white
```

---

## Anti-patterns

### ❌ Não usar `cn()` ao combinar classes condicionais
**Problema:** O desenvolvedor concatena strings de classes Tailwind com template literals ou operadores ternários diretamente no JSX.
**Por quê evitar:** Sem `twMerge`, classes conflitantes (e.g., `p-2 p-4`) não são resolvidas — o CSS aplica a última na folha de estilos (determinada pelo build), não a última no código. Isso produz bugs de estilo visuais e difíceis de rastrear.
**Solução:**
```tsx
// ❌ Errado — conflito de classes não resolvido
<div className={`p-2 ${isLarge ? 'p-4' : ''}`} />

// ✅ Correto — cn() via twMerge resolve o conflito, mantendo p-4
import { cn } from '@/lib/utils';
<div className={cn('p-2', isLarge && 'p-4')} />
// → resultado: 'p-4' (conflito resolvido corretamente)
```

---

### ❌ Hardcodar cores hexadecimais com valores arbitrários
**Problema:** O desenvolvedor usa classes como `text-[#1a2b3c]` ou `bg-[#ff6b35]` diretamente no JSX em vez de design tokens.
**Por quê evitar:** Cores hardcodadas impossibilitam tematização (dark mode, white-label), criam inconsistência visual com o design system e não respondem a mudanças globais de paleta. Também não funcionam com o mecanismo `dark:` de CSS variables.
**Solução:**
```tsx
// ❌ Errado — cor hardcodada, não suporta dark mode
<p className="text-[#1a2b3c]">Conteúdo</p>

// ✅ Correto — usa token do design system via CSS variable
// tailwind.config.ts: colors: { brand: { DEFAULT: 'hsl(var(--brand))' } }
<p className="text-brand">Conteúdo</p>

// ✅ Ou usa token semântico shadcn/ui
<p className="text-foreground">Conteúdo</p>
```

---

### ❌ Duplicar strings longas de classes Tailwind entre componentes
**Problema:** O desenvolvedor copia e cola a mesma string de 20+ classes Tailwind em múltiplos componentes em vez de extrair.
**Por quê evitar:** Manutenção distribuída — uma mudança de estilo precisa ser replicada em todos os lugares manualmente. Uma alteração esquecida introduz inconsistência visual silenciosa.
**Solução:**
```tsx
// ❌ Errado — mesma string duplicada em 5 lugares
<div className="rounded-lg border border-border bg-card p-4 shadow-sm hover:shadow-md transition-shadow" />

// ✅ Correto — extrai com cva para variantes reutilizáveis
import { cva } from 'class-variance-authority';

const card = cva('rounded-lg border border-border bg-card p-4 shadow-sm transition-shadow', {
  variants: {
    hoverable: { true: 'hover:shadow-md', false: '' },
  },
  defaultVariants: { hoverable: false },
});

<div className={card({ hoverable: true })} />
```

---

### ❌ Usar `!important` via prefixo `!` para sobrescrever estilos
**Problema:** O desenvolvedor usa o modificador `!` do Tailwind (ex: `!p-0`, `!text-red-500`) para forçar estilos em vez de resolver a especificidade.
**Por quê evitar:** `!important` cria uma corrida de especificidade — outros desenvolvedores precisam usar ainda mais `!important` para sobrescrever, tornando o CSS progressivamente impossível de manter.
**Solução:**
```tsx
// ❌ Errado — força override com !important
<Dialog className="!p-0 !rounded-none" />

// ✅ Correto — use cn() para sobrescrever na ordem correta
// O componente deve aceitar className e usar cn() internamente
const Dialog = ({ className, ...props }) => (
  <div className={cn('p-4 rounded-lg', className)} {...props} />
);

// Agora o caller pode sobrescrever normalmente:
<Dialog className="p-0 rounded-none" />
```

---

### ❌ Não usar `group` e `peer` para estados de pai/irmão
**Problema:** O desenvolvedor usa JavaScript (state + condicional) para aplicar estilos em um elemento filho quando o pai é hovered, em vez de usar o modificador `group` do Tailwind.
**Por quê evitar:** Estado em JS para hover puro é overhead desnecessário — re-renders, event listeners e complexidade de código para algo que CSS resolve nativamente com zero JS.
**Solução:**
```tsx
// ❌ Errado — hover gerenciado via React state
const [hovered, setHovered] = useState(false);
<div onMouseEnter={() => setHovered(true)} onMouseLeave={() => setHovered(false)}>
  <span className={hovered ? 'text-primary' : 'text-muted-foreground'}>Label</span>
</div>

// ✅ Correto — group/peer, zero JavaScript
<div className="group cursor-pointer rounded-lg p-4 hover:bg-accent">
  <span className="text-muted-foreground group-hover:text-primary transition-colors">
    Label
  </span>
</div>

// ✅ peer para estado de irmão (ex: input + label)
<input id="email" className="peer" type="email" />
<label htmlFor="email" className="text-sm peer-focus:text-primary transition-colors">
  Email
</label>
```

---

### ❌ Não usar `@layer components` para padrões CSS complexos reutilizáveis
**Problema:** O desenvolvedor repete padrões CSS complexos como strings Tailwind longas ou usa `<style>` scoped ad-hoc quando deveria definir uma classe de componente via `@layer components`.
**Por quê evitar:** Classes utilitárias fora de `@layer` têm especificidade maior que as classes Tailwind, causando conflitos de ordem. `@layer components` integra corretamente com o sistema de cascade do Tailwind e é purgeável pelo content scanner.
**Solução:**
```css
/* ✅ globals.css — padrão reutilizável via @layer */
@layer components {
  .prose-card {
    @apply rounded-xl border border-border bg-card p-6 shadow-sm;
    @apply transition-shadow hover:shadow-md;
  }

  .form-label {
    @apply text-sm font-medium text-foreground leading-none peer-disabled:opacity-70;
  }
}
```
```tsx
{/* Uso limpo no JSX */}
<div className="prose-card">...</div>
```

---

### ❌ Valores arbitrários de espaçamento fora do design system
**Problema:** O desenvolvedor usa valores arbitrários como `mt-[13px]`, `w-[127px]` ou `gap-[7px]` quando deveria usar a escala de espaçamento do Tailwind.
**Por quê evitar:** Valores fora da grade de espaçamento quebram consistência visual (alinhamentos desalinhados, layouts irregulares) e impedem ajustes globais de densidade — uma mudança no design token não afeta os valores hardcodados.
**Solução:**
```tsx
// ❌ Errado — valor mágico fora da grade
<div className="mt-[13px] gap-[7px] w-[127px]" />

// ✅ Correto — usa a escala do Tailwind (múltiplos de 4px)
<div className="mt-3 gap-2 w-32" />

// Se o valor realmente precisa ser customizado, adicione ao design system:
// tailwind.config.ts
// theme: { extend: { spacing: { 13: '3.25rem' } } }
```

---

### ❌ Ausência do variant `dark:` em componentes que usam cores
**Problema:** O desenvolvedor define cores de fundo e texto sem a variante `dark:`, hardcodando aparência para um único tema.
**Por quê evitar:** Sem `dark:`, o componente fica visualmente quebrado no dark mode — texto legível no tema claro pode se tornar ilegível sobre fundo escuro, e vice-versa.
**Solução:**
```tsx
// ❌ Errado — só funciona no tema claro
<div className="bg-white text-gray-900 border-gray-200">

// ✅ Correto — usa tokens semânticos que respondem ao dark mode automaticamente
<div className="bg-background text-foreground border-border">

// ✅ Ou, se precisar de cor específica, inclua o variant dark:
<div className="bg-white text-gray-900 dark:bg-gray-900 dark:text-gray-100 dark:border-gray-700">
```

---

### ❌ Safelist incorreta no Tailwind config (classes removidas em produção)
**Problema:** O desenvolvedor gera classes Tailwind dinamicamente via concatenação de strings (ex: `text-${color}-500`) e não adiciona essas classes à `safelist` do `tailwind.config`.
**Por quê evitar:** O content scanner do Tailwind analisa strings literais — classes montadas dinamicamente não são detectadas e são removidas pelo PurgeCSS no build de produção, quebrando estilos em prod mas funcionando em dev (que não purga).
**Solução:**
```typescript
// ❌ Errado — classe dinâmica invisível ao scanner
const color = 'red';
<span className={`text-${color}-500`} /> // 'text-red-500' removida em prod!

// ✅ Correto — mapa de classes completas (scanner detecta strings literais)
const colorMap = {
  red: 'text-red-500',
  green: 'text-green-500',
  blue: 'text-blue-500',
} as const;
<span className={colorMap[color]} />

// ✅ Ou safelist com pattern no tailwind.config.ts
// safelist: [{ pattern: /^text-(red|green|blue)-500$/ }]
```

---

### ❌ Estilo inline para valores dinâmicos em vez de CSS variables
**Problema:** O desenvolvedor usa `style={{ color: dynamicColor, width: dynamicWidth }}` para aplicar valores dinâmicos calculados em runtime.
**Por quê evitar:** Estilos inline têm especificidade máxima (sobrepõem qualquer classe Tailwind), não são purgeáveis, geram re-renders por referência de objeto nova a cada render, e impedem o uso de `hover:`, `dark:` e outros modificadores Tailwind.
**Solução:**
```tsx
// ❌ Errado — inline style bloqueia modificadores e tem alta especificidade
<div style={{ backgroundColor: brandColor, width: `${progress}%` }} />

// ✅ Correto — CSS custom property + classe Tailwind
<div
  className="bg-[--brand-color] transition-[width]"
  style={{ '--brand-color': brandColor, width: `${progress}%` } as React.CSSProperties}
/>

// ✅ Para progresso, use a utilidade arbitrary de width com variável
<div
  className="h-2 rounded-full bg-primary transition-[width] duration-300"
  style={{ width: `${progress}%` }}
/>
```
