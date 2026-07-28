# Git Expert — Estratégia de Branching, Fluxos e Decisões

Deep-dive conceitual do workflow Git: modelo de branching, fluxos ponta-a-ponta,
Conventional Commits, SemVer, rebase vs merge, PR workflow e automação de release.

> Este arquivo cobre o **modelo e as decisões** (o "porquê" e o "quando").
> As **configurações concretas** (`.pre-commit-config.yaml`, `commitlint.config.js`,
> `release.config.mjs`, template de PR, Makefile, `.gitleaks.toml`) estão em `patterns.md`.
> Ferramentas e versões em `stack.md`. Checklist de revisão em `pr-review-checklist.md`.

## Índice

- [1. Modelo de Branching — GitFlow Adaptado (ADR-023)](#1-modelo-de-branching--gitflow-adaptado-adr-023)
- [2. Fluxos Completos](#2-fluxos-completos)
- [3. Conventional Commits — Anatomia](#3-conventional-commits--anatomia)
- [4. SemVer — Cálculo e Disciplina](#4-semver--calculo-e-disciplina)
- [5. Rebase vs Merge — Árvore de Decisão](#5-rebase-vs-merge--arvore-de-decisao)
- [6. PR Workflow Ponta-a-Ponta](#6-pr-workflow-ponta-a-ponta)
- [7. Automação de Release — Fluxo](#7-automacao-de-release--fluxo)
- [8. Cadeia de Hooks (conceitual)](#8-cadeia-de-hooks-conceitual)

---

## 1. Modelo de Branching — GitFlow Adaptado (ADR-023)

adota **GitFlow Adaptado**: duas branches permanentes e branches temporárias de
vida curta. Trunk-Based Development com feature flags é aceitável apenas quando o time
tem CI/CD maduro e o ADR do projeto autoriza explicitamente — o padrão é GitFlow.

### Branches permanentes

| Branch | Papel | Recebe merge de | Protegida |
|--------|-------|-----------------|-----------|
| `main` | Estado de produção. Todo commit é releasable e taggeado (SemVer). | `release/*`, `hotfix/*` | Sim — sem push direto, sem force |
| `develop` | Linha de integração do próximo release. | `feature/*`, `bugfix/*`, `release/*` (back-merge), `hotfix/*` (back-merge) | Sim — sem push direto, sem force |

### Branches temporárias

| Prefixo | Origem | Destino | Naming | Uso |
|---------|--------|---------|--------|-----|
| `feature/` | `develop` | `develop` | `feature/PLAT-123-oauth2-pkce` | Nova funcionalidade |
| `bugfix/` | `develop` | `develop` | `bugfix/PLAT-456-pix-timeout` | Correção antes do release |
| `release/` | `develop` | `main` + `develop` | `release/1.4.0` | Estabilização + bump de versão |
| `hotfix/` | `main` | `main` + `develop` | `hotfix/1.3.1-invoice-crash` | Correção urgente em produção |

Regras de naming (validadas por hook `validate-branch-name`, ver `patterns.md`):
- Sempre `kebab-case` após o work item.
- Sempre incluir o ID do work item quando existir (`feature/PLAT-123-...`).
- `release/` e `hotfix/` nomeiam a **versão alvo**, não o work item.

### Branch protection (ADR-028)

| Policy | Configuração |
|--------|--------------|
| Push direto em `main`/`develop`/`release/*` | Bloqueado (`no-commit-to-branch` local + branch policy no servidor) |
| `push --force` em `main`/`develop`/`release/*` | Proibido por policy e por segurança |
| Build validation no PR | Pipeline deve passar antes do merge |
| Commitlint no CI | Todos os commits do PR validados |
| Gitleaks no CI | Rodado no diff do PR além do hook local |
| Squash merge | Preferencial para `feature`→`develop` (histórico limpo) |
| Delete branch after merge | Habilitado por padrão |

---

## 2. Fluxos Completos

### Feature flow

```bash
git checkout develop && git pull --rebase origin develop
git checkout -b feature/PLAT-123-oauth2-pkce
#... commits atômicos, Conventional Commits...
git fetch origin && git rebase origin/develop # sincroniza antes de abrir PR
git push -u origin feature/PLAT-123-oauth2-pkce
# abre PR feature -> develop; CI valida; review; SQUASH merge
```

### Release flow (bump de versão)

```bash
git checkout -b release/1.4.0 develop
# apenas estabilização: bugfixes, docs, bump de versão, changelog. SEM novas features.
# bump SemVer determinado pelos commits acumulados desde a última tag (ver §4).
git commit -m "chore(release): 1.4.0"
# merge release -> main (com --no-ff, cria tag anotada v1.4.0)
git checkout main && git merge --no-ff release/1.4.0
git tag -a v1.4.0 -m "Release v1.4.0" && git push origin main --tags
# BACK-MERGE obrigatório release -> develop (não perder os fixes de estabilização)
git checkout develop && git merge --no-ff release/1.4.0
git branch -d release/1.4.0
```

### Hotfix flow

```bash
git checkout -b hotfix/1.3.1-invoice-crash main # nasce de main, não de develop
# fix mínimo e cirúrgico
git checkout main && git merge --no-ff hotfix/1.3.1-invoice-crash
git tag -a v1.3.1 -m "Hotfix v1.3.1" && git push origin main --tags
# BACK-MERGE obrigatório hotfix -> develop
git checkout develop && git merge --no-ff hotfix/1.3.1-invoice-crash
git branch -d hotfix/1.3.1-invoice-crash
```

> **Regra do back-merge**: `release/*` e `hotfix/*` sempre voltam para `develop`. Deletar
> a branch antes do merge completo em `main` **e** `develop` é anti-padrão (ADR-028).

---

## 3. Conventional Commits — Anatomia

Formato: `type(scope): description #work-item`, corpo opcional, footers opcionais.
Conventional Commits é **pré-requisito** de SemVer automático e changelog gerado.

```
<type>(<scope>): <description> #<work-item>

<body — o POR QUÊ, não o O QUÊ>

<footers>
Refs: #PLAT-123
BREAKING CHANGE: <descrição da quebra de contrato>
Assisted-by: Claude claude-sonnet-4-6
```

| type | Efeito SemVer | Uso |
|------|---------------|-----|
| `feat` | MINOR | Nova funcionalidade |
| `fix` | PATCH | Correção de bug |
| `perf` | PATCH | Melhoria de performance |
| `refactor` | nenhum | Refatoração sem mudança de comportamento |
| `docs` | nenhum | Documentação |
| `test` | nenhum | Testes |
| `build` | nenhum | Build system, dependências |
| `ci` | nenhum | Pipeline CI/CD |
| `chore` | nenhum | Tarefas de manutenção |
| `style` | nenhum | Formatação (sem lógica) |
| `revert` | contexto | Reversão de commit anterior |

**Breaking changes** (bump MAJOR): sufixo `!` no type (`feat(api)!:...`) **ou** footer
`BREAKING CHANGE:`. Ambos disparam MAJOR; usar os dois deixa a intenção explícita.

**Regras **:
- `description` no imperativo, minúscula, sem ponto final, ≤ 72 chars.
- Work item no título (`#PLAT-123`) sempre que existir — rastreabilidade inegociável.
- `Assisted-by:` obrigatório quando o diff foi gerado/majoramente escrito por AI (ADR-028).
Não necessário para typo/autocomplete de uma linha.
- Um commit = um concern. Múltiplos concerns → múltiplos commits atômicos.

---

## 4. SemVer — Cálculo e Disciplina

`MAJOR.MINOR.PATCH`. O bump é **determinado pelos commits** desde a última tag, nunca por
decisão manual arbitrária.

```
Maior tipo presente desde a última tag → bump:
algum BREAKING CHANGE / type! → MAJOR (X.0.0)
senão algum feat → MINOR (x.Y.0)
senão algum fix|perf → PATCH (x.y.Z)
senão (só docs/chore/etc.) → sem release
```

Cálculo manual da próxima versão a partir do histórico:

```bash
LAST=$(git describe --tags --abbrev=0)
git log "$LAST"..HEAD --pretty=format:"%s%n%b" # inspecionar types e BREAKING CHANGE
```

**Disciplina**:
- Tags sempre **anotadas** (`git tag -a v1.4.0 -m...`), nunca lightweight — carregam autor e data.
- Versão é **imutável**: nunca re-taggear `v1.2.0` após deletar (quebra a garantia SemVer).
- Pre-releases em `develop`: `1.4.0-rc.1`, `1.4.0-beta.2` — configurar `prerelease` no
release tool para não gerar releases estáveis fora de `main`.

---

## 5. Rebase vs Merge — Árvore de Decisão

**Golden rule**: nunca reescreva história de branch **compartilhada**. Rebase só em branch
que é sua e ainda não foi integrada por outros.

| Situação | Estratégia | Comando |
|----------|-----------|---------|
| Sincronizar feature branch pessoal com `develop` | **Rebase** | `git fetch origin && git rebase origin/develop` |
| Limpar commits WIP antes de abrir PR | **Rebase interativo** | `git rebase -i origin/develop` (squash/fixup/reword) |
| Integrar `feature` → `develop` | **Squash merge** | histórico linear, 1 commit por feature |
| Integrar `release`/`hotfix` → `main` | **Merge `--no-ff`** | preserva o ponto de release para tag |
| Desfazer algo já em branch compartilhada | **Revert** (nunca rebase/reset) | `git revert <hash>` |

```bash
# ✅ rebase seguro (branch pessoal, ainda não integrada)
git rebase origin/develop
git push --force-with-lease # nunca --force puro

# ❌ NUNCA em branch compartilhada
git rebase origin/develop && git push --force origin develop # reescreve história alheia
```

**Resolução de conflitos no rebase**:
```bash
git rebase origin/develop
# resolver arquivo a arquivo
git add <arquivo> && git rebase --continue
git rebase --abort # se a coisa complicar, aborta e reavalia
```

**Squash vs merge commit** — por que squash para features: cada feature vira **um** commit
Conventional na `develop`, o changelog fica limpo e o `git bisect` fica preciso. Merge
`--no-ff` só onde o ponto de junção importa (releases/hotfixes → tag).

---

## 6. PR Workflow Ponta-a-Ponta

1. **Preparar**: branch rebasada com `develop`, commits atômicos e válidos, hooks locais
passando (`pre-commit run --all-files`), sem secrets (`gitleaks detect --no-git`).
2. **Abrir PR**: título `type(scope): resumo #work-item`; descrição a partir do template
(`patterns.md` §4) — descrição, tipo de mudança, work item, AI-assisted, checklist, evidências.
3. **CI valida**: build, testes, commitlint em todos os commits, gitleaks no diff.
4. **Review**: aplicar `pr-review-checklist.md` na ordem (Segurança primeiro). Classificar
🔴 Bloqueador / 🟡 Melhoria / 🔵 Observação.
5. **Aprovação + merge**: `feature`/`bugfix` → **squash** para `develop`; delete branch após merge.
6. **Release**: acumulado em `develop` sobe via `release/*` → `main` com tag (§2).

Criação via CLI:
```bash
gh pr create --base develop --title "feat(auth): add OAuth2 PKCE #PLAT-123" --body-file pr.md
az repos pr create --source-branch feature/... --target-branch develop --title "..." --description "..."
```

---

## 7. Automação de Release — Fluxo

Decisão (ver também SKILL.md): **semantic-release** quando o projeto publica artefato
(npm/PyPI/Docker/GitHub Releases) e o release é 100% automatizado; **conventional-changelog**
quando é serviço interno com aprovação manual do changelog.

### semantic-release (full-auto, por push em `main`)

```
push main → analisa commits → calcula versão SemVer → gera CHANGELOG.md
→ cria tag anotada → cria GitHub/Azure Release → publica artefato (se configurado)
```
- Em `develop`: configurar canal `prerelease` (ex.: `1.4.0-rc.1`) para não gerar estável.
- Config concreta (`release.config.mjs`, workflow) em `patterns.md` §3 e §5.

### conventional-changelog (semi-auto, interno)

```bash
conventional-changelog -p conventionalcommits -i CHANGELOG.md -s -r 0 # gera/atualiza
# revisão humana do CHANGELOG
git tag -a v1.4.0 -m "Release v1.4.0" && git push origin v1.4.0
```

### Anatomia do CHANGELOG

Seções por impacto, na ordem: **Breaking Changes**, **Features**, **Bug Fixes**,
**Performance**, **Other**. Cada entrada referencia o commit e o work item. Release sem
`CHANGELOG.md` atualizado é anti-padrão (ADR-028).

---

## 8. Cadeia de Hooks (conceitual)

Hooks locais dão feedback rápido; o CI é a rede de segurança obrigatória — nunca confie só
no local (`--no-verify` só em emergência documentada). Config em `patterns.md` §1.

| Estágio | Quando dispara | O que roda | Budget |
|---------|----------------|-----------|--------|
| `pre-commit` | antes de gravar o commit | trailing-whitespace, end-of-file, check-yaml, large-files, `no-commit-to-branch`, gitleaks, lint por linguagem | < 5s |
| `commit-msg` | ao gravar a mensagem | commitlint (`@commitlint/config-conventional`) | instantâneo |
| `pre-push` | antes do push | validate-branch-name, testes rápidos opcionais | < 30s |

**Invariantes** (ADR-028):
- Toda `rev` de repo de hook **pinned** (sem `latest`, sem `@main`).
- `gitleaks` **não pode ser removido** — é requisito de segurança; se "causa problema",
corrige-se a regra, não se remove o hook.
- `no-commit-to-branch` para `main` e `develop` é obrigatório.

---

## ADRs Relacionados

- **ADR-023** — GitFlow Adaptado, branch naming, Conventional Commits, PR template, SemVer.
- **ADR-028** — Hooks (pre-commit), AI-attribution (`Assisted-by`), release automation, rollback, branch protection.
