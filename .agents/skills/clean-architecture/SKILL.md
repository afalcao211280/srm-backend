---
name: clean-architecture
description: >
Princípios de Clean Architecture para desenhar sistemas manuteníveis e testáveis.
Use SEMPRE ao projetar sistemas/módulos, definir camadas e fronteiras, direcionar
dependências, ou refatorar acoplamento. Padrão: handler → service → repository →
domain (dependências apontam pra dentro). Acionar em "arquitetura", "camadas",
"boundaries", "use case", "dependency inversion", "onde colocar esta lógica".
version: "1.0.0"
category: Crosscutting
keywords:
- clean-architecture
- layering
- dependency-inversion
- use-cases
- boundaries
- solid
- domain
---

# Clean Architecture — Princípios (Cross-cutting)

Base: Robert C. Martin. Aplicável a todas as skills de Backend (golang, java, nestjs, python).

## A Regra da Dependência

**Dependências de código apontam SEMPRE pra dentro.** Camadas internas não conhecem as externas.

```
domain (entidades, regras) ← service (use cases) ← repository/handler (adapters) ← frameworks
```

Mapeamento: `handler` → `service` → `repository` → `domain`. `domain` tem zero deps externas.

## Princípios (por impacto)

1. **Dependency Direction (CRÍTICO)** — deps só pra dentro; interfaces pertencem ao consumidor (service declara, repository implementa); estruturas simples cruzam fronteiras; zero ciclos; depender de abstrações estáveis.
2. **Entity Design (CRÍTICO)** — entidades só com regras de negócio; ignoram persistência; encapsulam invariantes; value objects p/ conceitos de domínio; modelo rico, não anêmico.
3. **Use Case Isolation (ALTO)** — cada use case tem uma razão pra mudar; orquestra entidades (não implementa regra); sem lógica de apresentação; deps explícitas no constructor; define a fronteira transacional.
4. **Component Cohesion (ALTO)** — a estrutura grita o domínio, não o framework; agrupar o que muda junto; não forçar clientes a depender de código não usado.
5. **Boundaries (MÉDIO-ALTO)** — humble object nas bordas; boundaries parciais quando separação total é prematura; `main` é plugin da aplicação; adiar decisão de framework/DB.
6. **Interface Adapters (MÉDIO)** — controllers finos; mappers traduzem entre camadas; gateways escondem sistemas externos; anti-corruption layer p/ externos.
7. **Framework Isolation (MÉDIO)** — domínio com zero deps de framework; ORM/web só na camada de infra; DI container na borda; logging atrás de interface de domínio.
8. **Testing Architecture (BAIXO-MÉDIO)** — testes são parte da arquitetura; desenhar p/ testabilidade; testar camada isolada; verificar boundaries com testes.

## Anti-padrões

- ❌ `domain` importando framework/ORM/handler
- ❌ Regra de negócio no controller/handler
- ❌ Modelo anêmico (só getters/setters, lógica espalhada)
- ❌ Interface no provedor (deve morar no consumidor)
- ❌ Ciclos de dependência entre pacotes
- ❌ ORM entity vazando pra camada de apresentação

## Checklist

- [ ] Dependências apontam pra dentro (handler→service→repo→domain)
- [ ] `domain` sem deps externas (só stdlib)
- [ ] Interfaces declaradas no consumidor
- [ ] Use case orquestra, não implementa regra
- [ ] Zero ciclos de import
- [ ] Framework/ORM isolados na infra
- [ ] Cada camada testável isoladamente

## Cross-references

- `golang-expert`, `java-expert`, `nestjs-expert`, `python-expert` — aplicação por stack
- `testing-expert` — testes por camada

> **Atenção**: usa Clean Architecture pragmática (Go não é Java). Sem agregados/DDD tático excessivo; entidades em `domain/`, use cases em `service/`.
