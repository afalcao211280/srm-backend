# PR Review Checklist — git-expert

Checklist completo para revisão de Pull Requests. Classificação:
- 🔴 **Bloqueador** — merge proibido até resolver
- 🟡 **Melhoria** — deve ser corrigido antes do merge (ou documentar a exceção)
- 🔵 **Observação** — informativo, não bloqueia

---

## 1. Segurança (OWASP / Shift-Left)

| # | Verificação | Nível |
|---|-------------|-------|
| S1 | Nenhum secret, API key, password, token ou connection string hardcoded (inclui `Secret`/`secret.yaml` com `data`/`stringData` no Git — Sonar `secrets:S6694`) | 🔴 |
| S2 | Variáveis sensíveis lidas de env vars ou secret manager | 🔴 |
| S3 | Input de usuário sanitizado/validado antes de usar em SQL, shell, HTML | 🔴 |
| S4 | Queries SQL parametrizadas (sem concatenação de string) | 🔴 |
| S5 | Autenticação e autorização checadas nos endpoints/handlers corretos | 🔴 |
| S6 | Dados sensíveis não aparecem em logs (`password`, `token`, `secret`) | 🔴 |
| S7 | Dependências novas sem CVE crítica ou alta conhecida (`snyk`, `dependabot`) | 🟡 |
| S8 | HTTPS/TLS em toda comunicação externa | 🟡 |
| S9 | Operações destrutivas têm soft-delete ou backup antes | 🟡 |
| S10 | Headers de segurança presentes em endpoints HTTP (Content-Security-Policy, X-Frame-Options) | 🔵 |

---

## 2. Qualidade de Código

| # | Verificação | Nível |
|---|-------------|-------|
| Q1 | Nenhum `TODO`, `FIXME`, `HACK` sem issue associada (`// TODO ABC-123: ...`) — Sonar `go:S1135` | 🟡 |
| Q2 | Complexidade cognitiva ≤15 por função (Sonar `go:S3776`; não confundir com ciclomática) | 🟡 |
| Q3 | Nenhuma função com mais de 30 linhas sem extrair helper (Sonar `go:S138`) | 🟡 |
| Q4 | Nenhum bloco de código duplicado (DRY) | 🟡 |
| Q5 | Type safety: sem `any` (TS), sem `interface{}` sem justificativa (Go), sem `object` (Java) | 🟡 |
| Q6 | Nenhum código morto (funções/variáveis não usadas) | 🟡 |
| Q7 | Erros tratados explicitamente (sem `_ = err`, sem `except: pass`, sem `.catch(() => {})` vazio) | 🔴 |
| Q8 | Nomes de variáveis/funções autodescritivos (sem abreviações obscuras) | 🔵 |
| Q9 | SOLID: responsabilidade única, inversão de dependência | 🔵 |
| Q10 | Linter e type-check passando (sem warnings suprimidos sem justificativa) | 🔴 |

---

## 3. Testes

| # | Verificação | Nível |
|---|-------------|-------|
| T1 | Cobertura ≥ 80% no diff (não só no projeto todo) | 🔴 |
| T2 | Caminho feliz testado | 🔴 |
| T3 | Pelo menos um caso de erro/exceção testado | 🟡 |
| T4 | Casos edge testados: `null`/`nil`, lista vazia, valores negativos, strings vazias | 🟡 |
| T5 | Testes determinísticos — sem `time.Sleep`, sem datas hardcoded, sem random sem seed | 🟡 |
| T6 | Mocks isolados corretamente — não vazar estado entre testes | 🟡 |
| T7 | Testes de integração com containers reais (não mocks de DB) quando relevante | 🟡 |
| T8 | Nenhum teste novo marcado como `skip` ou `pending` sem issue | 🟡 |
| T9 | Nomes de testes descrevem o comportamento esperado (não o nome da função) | 🔵 |

---

## 4. Performance

| # | Verificação | Nível |
|---|-------------|-------|
| P1 | Nenhum N+1 query (loop que chama DB por iteração) | 🔴 |
| P2 | Queries com filtros usam índices (verificar com EXPLAIN se alterou schema) | 🟡 |
| P3 | Nenhuma alocação desnecessária em hot paths (loops internos) | 🟡 |
| P4 | Paginação em listagens que podem crescer indefinidamente | 🟡 |
| P5 | Cache justificado quando dado é lido frequentemente com baixa mutação | 🔵 |
| P6 | Algoritmo O(n²) ou pior em coleções potencialmente grandes | 🟡 |

---

## 5. Documentação

| # | Verificação | Nível |
|---|-------------|-------|
| D1 | README atualizado se mudou estrutura, setup ou dependências | 🟡 |
| D2 | Comentários explicam o POR QUÊ (não o O QUÊ) | 🔵 |
| D3 | APIs públicas documentadas (OpenAPI, JSDoc, godoc, docstring) | 🟡 |
| D4 | Breaking changes documentados no commit footer e na PR description | 🔴 |
| D5 | Migrations têm comentário explicando o propósito (não apenas SQL) | 🔵 |
| D6 | Variáveis de env novas documentadas no `.env.example` | 🟡 |

---

## 6. Manutenibilidade

| # | Verificação | Nível |
|---|-------------|-------|
| M1 | Arquitetura respeitada (nenhuma camada chamando outra diretamente sem passar pelas intermediárias) | 🔴 |
| M2 | Nenhuma dependência cíclica introduzida | 🔴 |
| M3 | Funções com mais de 3 parâmetros usam struct/object (`Options`/`Params`) — Sonar `go:S107` | 🟡 |
| M4 | Constantes extraídas de magic numbers/strings | 🟡 |
| M5 | Formatação consistente com o restante do codebase | 🔵 |
| M6 | Nenhuma importação de libs não-aprovadas no ADR sem discussão | 🟡 |

---

## 7. Git / Processo

| # | Verificação | Nível |
|---|-------------|-------|
| G1 | Commits seguem Conventional Commits (`type(scope): description #ID`) | 🟡 |
| G2 | Work item referenciado em todos os commits | 🟡 |
| G3 | Branch rebasada com develop/main (sem conflitos de merge) | 🟡 |
| G4 | Sem commits WIP, temp, fixup não squashados | 🟡 |
| G5 | Arquivos gerados não commitados (`dist/`, `build/`, `*.min.js`, `*.pyc`) | 🔴 |
| G6 | `.env` real não commitado (apenas `.env.example`) | 🔴 |
| G7 | Se AI gerou código: `Assisted-by:` footer presente nos commits | 🟡 |
| G8 | Tamanho do PR razoável (≤ 400 linhas de diff para review efetivo) | 🔵 |

---

## Output de Review (formato sugerido)

```markdown
## Revisão de PR — #<número> <título>

### 🔴 Bloqueadores (resolver antes do merge)
- **[S1]** `src/config/database.go:45` — API key hardcoded. Usar variável de ambiente.
- **[Q7]** `internal/service/user.go:78` — Erro ignorado `_ = err`. Tratar o erro.

### 🟡 Melhorias (devem ser corrigidas ou documentadas)
- **[T3]** `internal/service/user_test.go` — Falta teste para caso de usuário não encontrado (404).
- **[P1]** `internal/repository/order.go:120-135` — Loop com query DB por iteração. Considerar batch query.

### 🔵 Observações (informativo)
- **[Q8]** `internal/handler/pmt.go:15` — Variável `pmt` poderia ser `payment` para maior clareza.
- **[D5]** Migration `0042_add_order_status.sql` poderia ter comentário explicando a migração de dados existentes.

### ✅ Bem feito
- Testes com coverage >80% no novo código
- Error handling consistente com `fmt.Errorf("%w", err)`
- AI attribution corretamente declarada nos commits
```

---

## Guia de Uso Agentic

Ao ser acionado para revisar um PR:

1. **Receber o diff**: via paste, ou `gh pr diff <número>` / `az repos pr show --id <id>`
2. **Executar checklist**: percorrer cada categoria na ordem (Security primeiro)
3. **Classificar**: 🔴 / 🟡 / 🔵 para cada problema encontrado
4. **Gerar relatório**: no formato de Output acima
5. **Resumo executivo**: "X bloqueadores, Y melhorias, Z observações" — aprovado / mudanças requisitadas
