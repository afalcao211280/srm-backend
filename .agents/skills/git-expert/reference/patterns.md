# Git Expert — Padrões e Configurações

## 1. Pre-commit Config por Stack

### Go

```yaml
#.pre-commit-config.yaml
repos:
- repo: https://github.com/pre-commit/pre-commit-hooks
rev: v4.6.0
hooks:
- id: trailing-whitespace
- id: end-of-file-fixer
- id: check-yaml
- id: check-json
- id: check-added-large-files
args: ['--maxkb=500']
- id: no-commit-to-branch
args: ['--branch', 'main', '--branch', 'develop']

- repo: https://github.com/gitleaks/gitleaks
rev: v8.18.4
hooks:
- id: gitleaks

- repo: https://github.com/dnephin/pre-commit-golang
rev: v0.5.1
hooks:
- id: go-fmt
- id: go-imports
- id: go-vet-mod
- id: golangci-lint

- repo: https://github.com/jorisroovers/gitlint
rev: v0.19.1
hooks:
- id: gitlint

- repo: https://github.com/alessandrojcm/commitlint-pre-commit-hook
rev: v9.13.0
hooks:
- id: commitlint
stages: [commit-msg]
additional_dependencies: ['@commitlint/config-conventional']
```

### Python

```yaml
#.pre-commit-config.yaml
repos:
- repo: https://github.com/pre-commit/pre-commit-hooks
rev: v4.6.0
hooks:
- id: trailing-whitespace
- id: end-of-file-fixer
- id: check-yaml
- id: check-json
- id: check-added-large-files
args: ['--maxkb=500']
- id: no-commit-to-branch
args: ['--branch', 'main', '--branch', 'develop']

- repo: https://github.com/gitleaks/gitleaks
rev: v8.18.4
hooks:
- id: gitleaks

- repo: https://github.com/astral-sh/ruff-pre-commit
rev: v0.4.4
hooks:
- id: ruff
args: [--fix]
- id: ruff-format

- repo: https://github.com/alessandrojcm/commitlint-pre-commit-hook
rev: v9.13.0
hooks:
- id: commitlint
stages: [commit-msg]
additional_dependencies: ['@commitlint/config-conventional']
```

### Node/TypeScript (com Husky)

```bash
# Instalar
npm install --save-dev husky @commitlint/cli @commitlint/config-conventional lint-staged

# Inicializar husky
npx husky init
```

```bash
#.husky/pre-commit
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"
npx lint-staged
```

```bash
#.husky/commit-msg
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"
npx --no -- commitlint --edit ${1}
```

```js
// package.json (lint-staged section)
{
"lint-staged": {
"*.{ts,tsx,js,jsx}": ["eslint --fix", "prettier --write"],
"*.{json,yaml,yml,md}": ["prettier --write"]
}
}
```

---

## 2. Commitlint Completo

```js
// commitlint.config.js (ou.commitlintrc.js)
export default {
extends: ['@commitlint/config-conventional'],
rules: {
'scope-empty': [1, 'never'], // warn: escopo recomendado
'subject-max-length': [2, 'always', 72],
'body-max-line-length': [2, 'always', 100],
'footer-max-line-length': [2, 'always', 100],
'type-enum': [2, 'always', [
'feat', 'fix', 'docs', 'style', 'refactor',
'test', 'chore', 'ci', 'build', 'perf', 'revert'
]],
'footer-leading-blank': [2, 'always'],
'body-leading-blank': [2, 'always'],
},
helpUrl: 'https://www.conventionalcommits.org/',
};
```

---

## 3. Semantic-Release Config (release.config.mjs)

```js
// release.config.mjs — para projetos com publicação automática
export default {
branches: [
'main',
{ name: 'develop', prerelease: true },
{ name: 'release/v+([0-9])?(.{+([0-9]),x}).x', channel: 'release' },
],
plugins: [
['@semantic-release/commit-analyzer', {
preset: 'conventionalcommits',
releaseRules: [
{ type: 'feat', release: 'minor' },
{ type: 'fix', release: 'patch' },
{ type: 'perf', release: 'patch' },
{ type: 'refactor', release: false },
{ breaking: true, release: 'major' },
],
}],
['@semantic-release/release-notes-generator', {
preset: 'conventionalcommits',
presetConfig: {
types: [
{ type: 'feat', section: '✨ Features' },
{ type: 'fix', section: '🐛 Bug Fixes' },
{ type: 'perf', section: '⚡ Performance Improvements' },
{ type: 'refactor', section: '♻️ Code Refactoring', hidden: true },
{ type: 'docs', section: '📝 Documentation', hidden: true },
{ type: 'ci', section: '👷 CI/CD', hidden: true },
],
},
}],
['@semantic-release/changelog', {
changelogFile: 'CHANGELOG.md',
}],
'@semantic-release/npm', // remover se não for npm
['@semantic-release/git', {
assets: ['CHANGELOG.md', 'package.json'],
message: 'chore(release): ${nextRelease.version} [skip ci]\n\n${nextRelease.notes}',
}],
['@semantic-release/github', {
addReleases: 'bottom',
}],
],
};
```

---

## 4. Template de PR Description

```markdown
## 📋 Descrição

<!-- Descreva o QUE mudou e o POR QUÊ. O código mostra o COMO. -->
[Descrição clara e objetiva da mudança]

## 🏷️ Tipo de Mudança

- [ ] `feat` — Nova funcionalidade
- [ ] `fix` — Correção de bug
- [ ] `refactor` — Refatoração (sem nova feature nem fix)
- [ ] `docs` — Apenas documentação
- [ ] `test` — Adicionar/corrigir testes
- [ ] `ci` — Pipeline / infra
- [ ] `perf` — Melhoria de performance
- [ ] ⚠️ **BREAKING CHANGE** — Mudança não retrocompatível

## 🔗 Work Item

Relates to #<!-- ID do Work Item no Azure DevOps -->

## 🤖 AI-Assisted

<!-- Preencher se o diff foi gerado com apoio de AI -->
- [ ] Código gerado/revisado com AI (`Assisted-by:` nos commits)
- Modelo utilizado: <!-- ex: Claude claude-sonnet-4-6 -->

## ✅ Checklist

- [ ] Commits seguem Conventional Commits (`type(scope): description #ID`)
- [ ] Testes unitários adicionados/atualizados (cobertura ≥ 80%)
- [ ] Sem secrets ou credenciais no código (`gitleaks detect`)
- [ ] `lint` e `type-check` passando localmente
- [ ] Documentação atualizada (se necessário)
- [ ] Breaking changes documentados no corpo do commit
- [ ] PR rebasado com a branch de destino

## 📸 Screenshots / Evidence

<!-- Opcional: screenshots, logs relevantes, output de testes -->
```

---

## 5. Scripts de Release (Makefile targets)

```makefile
# Makefile
.PHONY: release-preview release-dry-run changelog

## Preview da próxima versão e changelog (sem publicar)
release-preview:
	npx semantic-release --dry-run

## Gerar apenas o CHANGELOG sem criar release
changelog:
	conventional-changelog -p conventionalcommits -i CHANGELOG.md -s -r 0

## Verificar commits desde a última tag
commits-since-tag:
	git log $(shell git describe --tags --abbrev=0)..HEAD --pretty=format:"%h %s"

## Calcular próximo SemVer manualmente (sem semantic-release)
next-version:
	@echo "Commits since last tag:"
	@git log $(shell git describe --tags --abbrev=0)..HEAD --pretty=format:"%s" | head -20
	@echo "\nLast tag: $(shell git describe --tags --abbrev=0)"
```

---

## 6. Gitleaks Config (`.gitleaks.toml`)

```toml
#.gitleaks.toml
title = " Gitleaks Config"

[extend]
useDefault = true # Herda regras padrão

[[rules]]
id = "api-key"
description = "Internal API Key"
regex = '''[_-]?api[_-]?key[_-]?[=:]["']?[A-Za-z0-9]{32,}'''
tags = ["key", "example"]

[[rules]]
id = "azure-connection-string"
description = "Azure Connection String"
regex = '''DefaultEndpointsProtocol=https;AccountName=[^;]+;AccountKey=[A-Za-z0-9+/=]{88}'''
tags = ["azure", "connection-string"]

[allowlist]
description = "Global Allowlist"
regexes = [
'''EXAMPLE_.*''', # Variáveis de exemplo explícitas
'''placeholder''',
'''your[_-]?api[_-]?key''',
]
paths = [
'''\.env\.example''', # Arquivo de exemplo (não real)
'''\.env\.sample''',
'''docs/''', # Documentação com exemplos
]
```

---

## 7. Validate Branch Name (`.validate-branch-namerc.json`)

```json
{
"pattern": "^(main|develop)$|^(feature|hotfix|release|task)/[A-Z]+-[0-9]+-[a-z0-9-]+$",
"errorMsg": "Branch name must follow GitFlow pattern: feature/PLAT-123-description or hotfix/PLAT-456-fix"
}
```

---

## 8. Git Config Recomendado (por projeto)

```bash
#.gitconfig local do projeto (via git config --local)

# Assinar commits (se GPG configurado)
git config --local commit.gpgsign true

# Rebase por padrão no pull
git config --local pull.rebase true

# Prune automático de branches remotas deletadas
git config --local fetch.prune true

# Push apenas a branch atual
git config --local push.default current

# Template de commit message
git config --local commit.template.gitmessage

# Editor para mensagens longas
git config --local core.editor "code --wait"
```

```
#.gitmessage (template de commit)
# <tipo>(<escopo>): <descrição> #<work-item>
#
# [corpo opcional: explique o POR QUÊ]
#
# Refs: #
# Assisted-by: (preencher se AI gerou o diff)
#
# Tipos: feat, fix, docs, style, refactor, test, chore, ci, build, perf, revert
# Escopo: nome do módulo/área afetada
# Máximo 72 chars no título
```

---

## Anti-patterns

### ❌ Commitar secrets no repositório
**Problema:** Incluir senhas, tokens, chaves de API ou certificados em arquivos comitados.
**Por quê evitar:** O histórico do git é permanente — mesmo após remoção, o secret persiste em `git log`. Rotacionar é obrigatório (não opcional) após qualquer exposição.
**Solução:**
```bash
# Prevenir com gitleaks no pre-commit
#.pre-commit-config.yaml
repos:
- repo: https://github.com/gitleaks/gitleaks
rev: v8.21.2
hooks:
- id: gitleaks

# Se já commitado: rotacione o secret IMEDIATAMENTE, depois use git filter-repo para limpar histórico
git filter-repo --path arquivo-com-secret --invert-paths
```

### ❌ Commits gigantes e monolíticos
**Problema:** Acumular muitas mudanças e fazer um único commit de centenas de arquivos.
**Por quê evitar:** `git bisect` fica ineficaz; code review é inviável; reverter o commit desfaz trabalho não relacionado.
**Solução:**
```bash
# Stage seletivo por hunk
git add -p # seleciona partes específicas de arquivos

# Commit atômico por contexto lógico
git commit -m "feat(auth): add JWT refresh token endpoint"
git commit -m "test(auth): add integration tests for refresh token"
```

### ❌ git push --force em branches compartilhadas
**Problema:** Usar `git push --force` ou `git push --force-with-lease` em branches que outros desenvolvedores já fizeram pull.
**Por quê evitar:** Reescreve o histórico que outros têm localmente — o colega faz `git pull` e tem conflitos de histórico divergente ou perde commits.
**Solução:**
```bash
# Em feature branches pessoais: --force-with-lease é aceitável
git push --force-with-lease origin feature/minha-feature

# Em develop/main/release: NUNCA force push
# Se necessário desfazer: use git revert (cria commit de desfazimento)
git revert HEAD~3..HEAD # desfaz os últimos 3 commits sem reescrever histórico
```

### ❌ Não usar Conventional Commits
**Problema:** Escrever mensagens de commit livres: "ajustes", "fix", "WIP", "tentando de novo".
**Por quê evitar:** Impossibilita geração automática de CHANGELOG, SemVer automation com semantic-release e rastreabilidade de features/fixes por versão.
**Solução:**
```bash
# Formato obrigatório
feat(users): add email verification on signup
fix(payments): prevent double charge on network retry
chore(deps): bump golang.org/x/net to v0.38.0
# Instale commitlint para enforçar no CI
```

### ❌ Feature branches de longa duração
**Problema:** Manter branches de feature abertas por semanas sem sincronizar com main/develop.
**Por quê evitar:** Divergência acumula — merge conflicts crescem exponencialmente com o tempo; integração contínua se torna "integração ocasional".
**Solução:**
```bash
# Rebase diário na feature branch para manter atualizado
git fetch origin
git rebase origin/develop

# Ou use Trunk-Based Development com feature flags para merges pequenos e frequentes
```

### ❌ Commitar diretamente em main/master
**Problema:** Fazer commits ou pushes diretamente na branch principal sem abrir PR.
**Por quê evitar:** Bypassa code review, CI e quality gates — bugs vão direto para produção; sem rastreabilidade de decisões.
**Solução:**
```bash
# Configure branch protection rules no GitHub/Azure DevOps
# Exija: 1+ reviewers, CI passing, no direct push
# Trabalhe SEMPRE em feature branches, abra PR para merge
```

### ❌ Mensagens de commit sem contexto
**Problema:** Escrever `git commit -m "fix"` ou `git commit -m "alterações de hoje"`.
**Por quê evitar:** Em 6 meses, ninguém (incluindo você) saberá por que aquela mudança foi feita — rastrear bugs com `git blame` e `git log` fica impossível.
**Solução:**
```bash
# Responda: O QUÊ e POR QUÊ (não o como — o código já mostra o como)
git commit -m "fix(checkout): prevent null pointer when cart is empty

Quando o usuário acessa /checkout sem itens no carrinho, o campo
discount era acessado antes do cart ser validado. Adicionada
validação antecipada com retorno de erro 422.

Closes #347"
```
