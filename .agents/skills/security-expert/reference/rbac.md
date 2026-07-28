# Controle de Acesso — RBAC + ABAC + OPA

O **baseline** (`security-baseline.md`) já exige: RBAC com least privilege, permissão verificada em CADA endpoint, deny-by-default, IDOR (ownership), mass assignment (allowlist de campos). Este arquivo detalha **modelagem, ABAC, policy engine e auditoria** — alinhado ao ADR-019 (RBAC+ABAC com OPA).

## Modelagem RBAC

- **Permission** = `resource:action` (ex: `expense:approve`, `user:read`). Granular o suficiente para least privilege.
- **Role** = conjunto de permissions + herança de outras roles (`inheritFrom`). Resolver permissões recursivamente: `editor` herda `viewer` e adiciona as próprias.
- **Hierarquia evita role explosion**: sem herança, N atributos × M ações geram produto cartesiano de roles (`user_admin`, `user_viewer`, `user_editor`...). Compor por herança em vez de duplicar.
- **Wildcard** (`*:*`) para admin é atalho conveniente mas perigoso — preferir modelar `allow if role == "admin"` na policy a um wildcard que bypassa toda checagem.

## ABAC — quando RBAC não basta

RBAC é estático (role → permissão). ABAC decide por **atributos em tempo de request**:

- **Subject**: id, departamento, nível, role do usuário.
- **Resource**: owner_id, departamento, classificação.
- **Action**: read/write/approve.
- **Environment**: horário, janela de manutenção, IP/rede.

Casos que só ABAC cobre bem:
- **Ownership** (`user.id == resource.owner_id`) — IDOR modelado como policy.
- **Separation of duties** (`user.department == resource.department`).
- **Temporal** (acesso só em horário comercial / fora de janela de manutenção).

**Precedência: DENY vence.** Avaliar todas as policies; um DENY é veto sobre qualquer ALLOW. Trade-off: RBAC = simples/estático; ABAC = flexível/overhead. Usar RBAC como base e ABAC para as regras dinâmicas.

## Policy engine (OPA — padrão ADR-019)

- **OPA (Open Policy Agent)** desacopla a policy da aplicação: regras em Rego, app só consulta `allow`. Para microsserviços (policy-as-code, versionável).
- Mapeamento direto de ABAC para Rego:

```rego
allow {
    input.user.id == input.resource.owner_id
    input.action == "read"
}
deny { input.system.maintenance_mode == true }
```

- **Casbin** é alternativa para monolito; **Cedar** outra opção.

## Padrões de enforcement

- **Middleware/decorator**: `authorize(resource, action)` no router — separa a lógica (no engine) da aplicação (no handler). Reutilizável.
- **Inline policy-as-code**: Spring `@PreAuthorize("hasRole('ADMIN') or #id == authentication.principal.id")` (SpEL com variáveis de rota) — precursor de policy engine, migrável para OPA.
- **Guard customizado** (em service/bean) quando a lógica combina múltiplos atributos (role + departamento) — ponte entre framework RBAC e regras ABAC.
- **Nunca confiar em authz no cliente** — toda decisão no servidor.

## Auditoria e change management

- **Metadata** em roles/permissions: `created_at`, `modified_by`, `revision`.
- **Audit trail**: log de atribuição/revogação de role e de mudança de policy (quem, quando, o quê) — `role_assignment_log`, `policy_change_log`.
- **Access reviews** periódicos: revalidar quem tem cada role.
- **Approval workflow** para elevação de privilégio.

## Checklist de access control

- [ ] Hierarquia de roles auditável (sem explosão)
- [ ] Mudanças de permissão logadas
- [ ] Access reviews agendados
- [ ] Policies ABAC testadas
- [ ] Policy engine (OPA/Casbin) integrado em microsserviços
- [ ] Separation of duties aplicada
- [ ] Elevação de privilégio controlada (approval + log)
