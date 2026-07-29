# ADR-014 — GitHub Flow em dois repos

**Decisão:** GitHub Flow — `main` sempre liberável, uma branch curta por feature (`feature/<escopo>`), merge via Pull Request, histórico linear por rebase, tags semver.

**Topologia:** backend e frontend são repositórios independentes. Nenhum PR atravessa os dois.
