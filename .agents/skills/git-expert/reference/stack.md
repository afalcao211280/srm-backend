# Git Expert — Stack Canônica

## Ferramentas e Versões

| Categoria | Ferramenta | Versão | Link Oficial | Notas |
|-----------|-----------|--------|-------------|-------|
| **Hooks** | | | | |
| Hook framework | pre-commit | 3.7+ | https://pre-commit.com | Language-agnostic, pinned versions via .pre-commit-config.yaml |
| Hook manager (JS) | husky | 9.x+ | https://typicode.github.io/husky | Alternativa para projetos Node/TS; integra com package.json |
| **Commit Validation** | | | | |
| Lint de commits | commitlint | 19.x+ | https://commitlint.js.org | `@commitlint/config-conventional` como base |
| Secret detection | gitleaks | 8.x+ | https://github.com/gitleaks/gitleaks | Hook pre-commit; standalone CLI |
| Branch naming | validate-branch-name | 1.3+ | https://github.com/nickytonline/validate-branch-name | Pattern via `.validate-branch-namerc.json` |
| **Release Automation** | | | | |
| Release (full-auto) | semantic-release | 24.x+ | https://semantic-release.gitbook.io | Plugins por ecosystem (npm, pypi, docker) |
| Changelog (manual) | conventional-changelog | 8.x+ | https://github.com/conventional-changelog | CLI: `conventional-changelog -p angular -i CHANGELOG.md -s` |
| **CLI de PRs** | | | | |
| GitHub PRs | GitHub CLI (`gh`) | 2.50+ | https://cli.github.com | `gh pr create`, `gh pr review`, `gh pr merge` |
| Azure DevOps PRs | Azure CLI (`az repos`) | 2.60+ | https://learn.microsoft.com/cli/azure | `az repos pr create`, `az repos pr list` |
| **Linting por linguagem (hooks)** | | | | |
| Go | golangci-lint | 1.57+ | https://golangci-lint.run | Hook pre-commit: `golangci-lint run` |
| Go format | gofmt / goimports | stdlib | — | Sempre aplicar antes de commit |
| Python | ruff | 0.4+ | https://docs.astral.sh/ruff | Substitui flake8 + isort + black em um |
| Python format | black | 24.x+ | https://black.readthedocs.io | `black --check` no hook |
| Node/TS | ESLint | 9.x+ | https://eslint.org | Flat config (`eslint.config.js`) |
| Node/TS format | Prettier | 3.x+ | https://prettier.io | `prettier --check` no hook |
| Java | Checkstyle | 10.x+ | https://checkstyle.org | Via Maven/Gradle plugin |
| IaC | TFLint | 0.50+ | https://github.com/terraform-linters/tflint | Hook para Terraform |
| Docker | hadolint | 2.x+ | https://github.com/hadolint/hadolint | Lint de Dockerfile |
| Shell | shellcheck | 0.10+ | https://github.com/koalaman/shellcheck | Lint de .sh |
| YAML | yamllint | 1.35+ | https://github.com/adrienverge/yamllint | Lint de YAML (pipelines, configs) |
| Markdown | markdownlint | 0.39+ | https://github.com/DavidAnson/markdownlint | Lint de docs |
| **Assinatura de Commits** | | | | |
| GPG signing | gnupg | 2.4+ | https://gnupg.org | `git config commit.gpgsign true` |
| Sigstore | cosign | 2.x+ | https://docs.sigstore.dev | Signing de artefatos (containers, releases) |

## Semantic-Release Plugins por Ecosystem

| Ecosystem | Plugin | Output |
|-----------|--------|--------|
| npm | `@semantic-release/npm` | `package.json` version bump + npm publish |
| Python (PyPI) | `semantic-release-pypi` | `pyproject.toml` + PyPI publish |
| Docker | `@semantic-release/exec` (custom) | Docker tag + push |
| Go | `@semantic-release/exec` (custom) | Git tag + GitHub Release |
| GitHub Release | `@semantic-release/github` | Release notes + assets |
| Changelog | `@semantic-release/changelog` | `CHANGELOG.md` atualizado |
| Git | `@semantic-release/git` | Commit `package.json` + `CHANGELOG.md` |

## SemVer: Conventional Commits → Bump

| Commit Type | SemVer Bump | Exemplo |
|-------------|-------------|---------|
| `fix:` | PATCH (x.x.X) | `fix(auth): correct token expiry check` |
| `perf:` | PATCH | `perf(db): add index on user email` |
| `feat:` | MINOR (x.X.0) | `feat(payment): add PIX support` |
| `feat!:` ou `BREAKING CHANGE:` | MAJOR (X.0.0) | `feat(api)!: remove v1 endpoints` |
| `docs:`, `style:`, `refactor:`, `test:`, `chore:`, `ci:` | nenhum | Não gera release |

## Commitlint Config de Referência

```js
// commitlint.config.js
export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'scope-empty': [1, 'never'],           // warn se sem escopo
    'subject-max-length': [2, 'always', 72],
    'body-max-line-length': [2, 'always', 100],
    'footer-max-line-length': [2, 'always', 100],
    'type-enum': [2, 'always', [
      'feat', 'fix', 'docs', 'style', 'refactor',
      'test', 'chore', 'ci', 'build', 'perf', 'revert'
    ]],
  },
};
```
