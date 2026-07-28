# React Expert — Padrões de Código

## 1. Componente Funcional Padrão

```tsx
// src/components/features/user/UserCard.tsx
import { type FC } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import type { User } from '@/types/user';

interface UserCardProps {
user: User;
onEdit?: (id: string) => void;
}

export const UserCard: FC<UserCardProps> = ({ user, onEdit }) => {
return (
<div className="flex items-center gap-4 rounded-lg border p-4">
<Avatar>
<AvatarImage src={user.avatarUrl} alt={user.name} />
<AvatarFallback>{user.name[0]}</AvatarFallback>
</Avatar>
<div className="flex-1">
<p className="font-medium">{user.name}</p>
<p className="text-sm text-muted-foreground">{user.email}</p>
</div>
<Badge variant={user.active? 'default': 'secondary'}>
{user.active? 'Ativo': 'Inativo'}
</Badge>
{onEdit && (
<button onClick={() => onEdit(user.id)} className="text-sm text-primary hover:underline">
Editar
</button>
)}
</div>
);
};
```

---

## 2. Custom Hook com TanStack Query

```tsx
// src/hooks/useUsers.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { userApi } from '@/api/users';
import type { UserCreate } from '@/types/user';

export const USER_KEYS = {
all: ['users'] as const,
list: () => [...USER_KEYS.all, 'list'] as const,
detail: (id: string) => [...USER_KEYS.all, id] as const,
};

export function useUsers() {
return useQuery({
queryKey: USER_KEYS.list(),
queryFn: userApi.list,
staleTime: 5 * 60 * 1000, // 5 min
});
}

export function useUser(id: string) {
return useQuery({
queryKey: USER_KEYS.detail(id),
queryFn: () => userApi.getById(id),
enabled: Boolean(id),
});
}

export function useCreateUser() {
const queryClient = useQueryClient();
return useMutation({
mutationFn: (data: UserCreate) => userApi.create(data),
onSuccess: () => {
queryClient.invalidateQueries({ queryKey: USER_KEYS.list() });
},
});
}
```

---

## 3. Form com React Hook Form + Zod

```tsx
// src/components/features/user/CreateUserForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Form, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { useCreateUser } from '@/hooks/useUsers';

const createUserSchema = z.object({
name: z.string().min(2, 'Nome deve ter ao menos 2 caracteres').max(100),
email: z.string().email('Email inválido'),
password: z.string().min(8, 'Senha deve ter ao menos 8 caracteres'),
});

type CreateUserForm = z.infer<typeof createUserSchema>;

export function CreateUserForm() {
const createUser = useCreateUser();
const form = useForm<CreateUserForm>({
resolver: zodResolver(createUserSchema),
defaultValues: { name: '', email: '', password: '' },
});

const onSubmit = form.handleSubmit(async (data) => {
await createUser.mutateAsync(data);
form.reset();
});

return (
<Form {...form}>
<form onSubmit={onSubmit} className="space-y-4">
<FormField
control={form.control}
name="name"
render={({ field }) => (
<FormItem>
<FormLabel>Nome</FormLabel>
<Input {...field} placeholder="João Silva" />
<FormMessage />
</FormItem>
)}
/>
<FormField
control={form.control}
name="email"
render={({ field }) => (
<FormItem>
<FormLabel>Email</FormLabel>
<Input {...field} type="email" placeholder="joao@example.com" />
<FormMessage />
</FormItem>
)}
/>
<Button type="submit" disabled={createUser.isPending}>
{createUser.isPending? 'Criando...': 'Criar Usuário'}
</Button>
</form>
</Form>
);
}
```

---

## 4. Zustand Store

```tsx
// src/stores/authStore.ts
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types/user';

interface AuthState {
user: User | null;
token: string | null;
setAuth: (user: User, token: string) => void;
clearAuth: () => void;
isAuthenticated: () => boolean;
}

export const useAuthStore = create<AuthState>()(
persist(
(set, get) => ({
user: null,
token: null,
setAuth: (user, token) => set({ user, token }),
clearAuth: () => set({ user: null, token: null }),
isAuthenticated: () => Boolean(get().token),
}),
{ name: 'auth' },
),
);
```

---

## 5. Axios Instance com Interceptors

```tsx
// src/api/client.ts
import axios from 'axios';
import { useAuthStore } from '@/stores/authStore';

export const apiClient = axios.create({
baseURL: import.meta.env['VITE_API_URL'],
timeout: 10_000,
headers: { 'Content-Type': 'application/json' },
});

// Request: injetar token
apiClient.interceptors.request.use((config) => {
const token = useAuthStore.getState().token;
if (token) {
config.headers.Authorization = `Bearer ${token}`;
}
return config;
});

// Response: tratar 401
apiClient.interceptors.response.use(
(response) => response,
(error) => {
if (axios.isAxiosError(error) && error.response?.status === 401) {
useAuthStore.getState().clearAuth();
window.location.href = '/login';
}
return Promise.reject(error);
},
);
```

---

## 6. API Layer (separada de hooks)

```tsx
// src/api/users.ts
import { apiClient } from './client';
import type { User, UserCreate, UserListResponse } from '@/types/user';

export const userApi = {
list: async (): Promise<User[]> => {
const { data } = await apiClient.get<UserListResponse>('/v1/users');
return data.items;
},

getById: async (id: string): Promise<User> => {
const { data } = await apiClient.get<User>(`/v1/users/${id}`);
return data;
},

create: async (payload: UserCreate): Promise<User> => {
const { data } = await apiClient.post<User>('/v1/users', payload);
return data;
},

update: async (id: string, payload: Partial<UserCreate>): Promise<User> => {
const { data } = await apiClient.patch<User>(`/v1/users/${id}`, payload);
return data;
},

delete: async (id: string): Promise<void> => {
await apiClient.delete(`/v1/users/${id}`);
},
};
```

---

## 7. Teste com Vitest + Testing Library

```tsx
// src/components/features/user/UserCard.test.tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { UserCard } from './UserCard';

const mockUser = {
id: '1',
name: 'João Silva',
email: 'joao@example.com',
avatarUrl: '',
active: true,
};

describe('UserCard', () => {
it('renderiza nome e email do usuário', () => {
render(<UserCard user={mockUser} />);
expect(screen.getByText('João Silva')).toBeInTheDocument();
expect(screen.getByText('joao@example.com')).toBeInTheDocument();
});

it('chama onEdit quando botão de editar é clicado', async () => {
const onEdit = vi.fn();
render(<UserCard user={mockUser} onEdit={onEdit} />);
await userEvent.click(screen.getByRole('button', { name: /editar/i }));
expect(onEdit).toHaveBeenCalledWith('1');
});

it('não exibe botão de editar quando onEdit não fornecido', () => {
render(<UserCard user={mockUser} />);
expect(screen.queryByRole('button', { name: /editar/i })).not.toBeInTheDocument();
});
});
```

---

## Anti-patterns

### ❌ useEffect para buscar dados
**Problema:** O desenvolvedor usa `useEffect` + `useState` para fazer fetch de dados do servidor.
**Por quê evitar:** Causa race conditions, não lida com cache, loading states duplicados e revalidação. O React 19 e o TanStack Query resolvem isso de forma declarativa e segura.
**Solução:**
```tsx
// Errado
useEffect(() => {
fetch('/api/users').then(r => r.json()).then(setUsers);
}, []);

// Correto — use TanStack Query
const { data: users, isLoading } = useQuery({
queryKey: USER_KEYS.list(),
queryFn: userApi.list,
staleTime: 5 * 60 * 1000,
});
```

---

### ❌ Array de dependências ausente ou vazio incorretamente no useEffect
**Problema:** O desenvolvedor omite o array de dependências ou usa `[]` quando o efeito depende de valores externos.
**Por quê evitar:** Com array vazio, o efeito captura valores stale (closures desatualizadas). Sem array, o efeito roda em todo render causando loops infinitos ou requisições desnecessárias.
**Solução:**
```tsx
// Errado — userId fica stale
useEffect(() => {
fetchUser(userId); // userId nunca atualiza dentro do efeito
}, []);

// Correto — declare todas as dependências
useEffect(() => {
fetchUser(userId);
}, [userId]);
```

---

### ❌ Definir componentes dentro de outros componentes
**Problema:** O desenvolvedor declara um componente filho dentro do corpo do componente pai.
**Por quê evitar:** A cada render do pai, o React cria uma nova referência de função para o filho, destruindo e remontando o componente inteiro — perdendo estado e causando re-renders desnecessários.
**Solução:**
```tsx
// Errado
function ParentComponent() {
function ChildItem({ name }: { name: string }) { // recriado a cada render
return <li>{name}</li>;
}
return <ul>{items.map(i => <ChildItem key={i.id} name={i.name} />)}</ul>;
}

// Correto — declare fora
function ChildItem({ name }: { name: string }) {
return <li>{name}</li>;
}

function ParentComponent() {
return <ul>{items.map(i => <ChildItem key={i.id} name={i.name} />)}</ul>;
}
```

---

### ❌ Mutação direta do estado
**Problema:** O desenvolvedor modifica arrays ou objetos do estado diretamente com `push`, `splice` ou atribuição de propriedade.
**Por quê evitar:** O React compara referências para detectar mudanças. Mutação direta não cria uma nova referência, então o componente não re-renderiza e a UI fica dessincronizada do estado real.
**Solução:**
```tsx
// Errado
const [items, setItems] = useState<string[]>([]);
items.push('novo'); // não dispara re-render
setItems(items); // mesma referência, React ignora

// Correto — crie novo array/objeto
setItems(prev => [...prev, 'novo']);

// Correto para objetos
setUser(prev => ({...prev, name: 'João' }));
```

---

### ❌ Usar index de array como key em listas dinâmicas
**Problema:** O desenvolvedor usa o índice do array como `key` em listas que podem ser reordenadas, filtradas ou ter itens removidos.
**Por quê evitar:** O React usa a `key` para identificar quais elementos mudaram. Com índices, reordenar itens confunde o reconciliador, causando bugs de estado incorreto em inputs controlados, animações e componentes com estado interno.
**Solução:**
```tsx
// Errado
{users.map((user, index) => (
<UserCard key={index} user={user} /> // key muda quando lista é reordenada
))}

// Correto — use identificador estável e único
{users.map((user) => (
<UserCard key={user.id} user={user} />
))}
```

---

### ❌ Excesso de useCallback/useMemo (otimização prematura)
**Problema:** O desenvolvedor envolve todas as funções e valores em `useCallback`/`useMemo` sem medir se há problema de performance real.
**Por quê evitar:** Cada `useCallback`/`useMemo` tem custo de criação do closure e comparação de dependências a cada render. Aplicado desnecessariamente, o código fica mais complexo sem ganho — e pode ser mais lento que a versão simples.
**Solução:**
```tsx
// Errado — memoização desnecessária em componente simples
const handleClick = useCallback(() => {
setCount(c => c + 1);
}, []); // o componente não passa isso para filho memoizado

// Correto — use somente quando:
// 1. A função é passada como prop para um componente envolto em React.memo
// 2. A função é dependência de outro useEffect/useMemo
const stableCallback = useCallback(() => {
doExpensiveOperation(id);
}, [id]); // passado para <HeavyChild onClick={stableCallback} />
```

---

### ❌ Não limpar efeitos colaterais no useEffect
**Problema:** O desenvolvedor cria subscriptions, timers ou listeners no `useEffect` sem retornar uma função de cleanup.
**Por quê evitar:** Causa memory leaks — a subscription continua ativa mesmo após o componente ser desmontado, podendo chamar `setState` em componente inexistente e gerar erros em produção.
**Solução:**
```tsx
// Errado — sem cleanup
useEffect(() => {
const sub = eventBus.subscribe('update', handler);
// sub nunca é cancelada
}, []);

// Correto — retorne a função de cleanup
useEffect(() => {
const sub = eventBus.subscribe('update', handler);
return () => sub.unsubscribe();
}, [handler]);
```

---

### ❌ Chamar hooks condicionalmente ou dentro de loops
**Problema:** O desenvolvedor usa `if`, ternário ou `for` antes de uma chamada de hook.
**Por quê evitar:** As Rules of Hooks exigem que hooks sejam chamados na mesma ordem em todo render. Chamadas condicionais quebram essa ordem, causando erros difíceis de depurar e comportamento imprevisível do estado.
**Solução:**
```tsx
// Errado
function Component({ isAdmin }: { isAdmin: boolean }) {
if (isAdmin) {
const data = useAdminData(); // viola Rules of Hooks
}
}

// Correto — chame o hook incondicionalmente e controle dentro
function Component({ isAdmin }: { isAdmin: boolean }) {
const data = useAdminData(); // sempre chamado
if (!isAdmin) return null;
return <AdminPanel data={data} />;
}
```

---

### ❌ Prop drilling excessivo (3+ níveis)
**Problema:** O desenvolvedor passa props por 3 ou mais camadas de componentes intermediários que não usam os dados.
**Por quê evitar:** Cria acoplamento forte entre componentes, dificulta refatoração e obriga todos os componentes intermediários a conhecer dados que não são seus. A manutenção se torna exponencialmente custosa.
**Solução:**
```tsx
// Errado — user percorre 3 níveis que não precisam dele
<Page user={user}>
<Layout user={user}>
<Sidebar user={user}>
<UserAvatar user={user} /> {/* único que usa */}
</Sidebar>
</Layout>
</Page>

// Correto — use Zustand store ou Context focado
const { user } = useAuthStore(); // lido diretamente onde precisa

// Ou composition pattern
function Page() {
return <Layout sidebar={<UserAvatar />} />;
}
```

---

### ❌ staleTime: 0 em queries que não mudam frequentemente
**Problema:** O desenvolvedor deixa o `staleTime` padrão (0) no TanStack Query para dados que raramente mudam.
**Por quê evitar:** Com `staleTime: 0`, toda montagem de componente que usa a query dispara um refetch em background, gerando requisições desnecessárias ao servidor e flickering na UI quando o dado é substituído.
**Solução:**
```tsx
// Errado — refetch em todo mount mesmo para dados estáticos
useQuery({
queryKey: ['config'],
queryFn: fetchAppConfig,
// staleTime: 0 (padrão)
});

// Correto — defina staleTime adequado ao domínio
useQuery({
queryKey: ['config'],
queryFn: fetchAppConfig,
staleTime: 10 * 60 * 1000, // 10 min — configuração raramente muda
});

// No QueryClient global para defaults sensatos
new QueryClient({
defaultOptions: { queries: { staleTime: 60_000 } },
});
```
