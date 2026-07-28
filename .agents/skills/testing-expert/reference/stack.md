# Testing Expert — Stack Canônica

## Testing Strategy — Versão Canônica

| Categoria | Ferramenta | Versão | Link | Notas |
|-----------|-----------|--------|------|-------|
| **Node/TypeScript** | | | | |
| Unit/Integration | Vitest | 2.x | https://vitest.dev | Vite-native. Preferido para novos projetos Node |
| Unit/Integration | Jest | 29.x | https://jestjs.io | Padrão estabelecido. `@nestjs/testing` |
| Component (React) | React Testing Library | 16.x | https://testing-library.com/react | Comportamento > implementação |
| Component (Vue) | Vue Test Utils | 2.x | https://test-utils.vuejs.org | |
| Component (Angular) | Angular Testing Library | latest | https://testing-library.com/angular | Sobre TestBed |
| Mocks HTTP | MSW (Mock Service Worker) | 2.x | https://mswjs.io | Intercepta na rede, não na implementação |
| **Python** | | | | |
| Unit/Integration | pytest | 8.x | https://pytest.org | Fixtures, parametrize, plugins |
| Async | pytest-asyncio | 0.23+ | https://github.com/pytest-dev/pytest-asyncio | Para código async |
| Coverage | pytest-cov | 5.x | https://pytest-cov.readthedocs.io | `--cov --cov-report=xml` |
| Factories | factory_boy | 3.x | https://factoryboy.readthedocs.io | Fixtures como código |
| **Java** | | | | |
| Unit | JUnit 5 | 5.11+ | https://junit.org/junit5 | `@Test`, `@ParameterizedTest` |
| Mocks | Mockito | 5.x | https://site.mockito.org | `@Mock`, `@InjectMocks` |
| Spring | Spring Boot Test | 3.3+ | https://spring.io | `@SpringBootTest`, `@WebMvcTest` |
| REST API | RestAssured | 5.x | https://rest-assured.io | Fluent API testing |
| **Go** | | | | |
| Unit | testing (stdlib) | built-in | — | `t.Run`, `t.Parallel()` |
| Assertions | testify | 1.9+ | https://github.com/stretchr/testify | `assert`, `require`, `suite` |
| HTTP | httptest | built-in | — | `httptest.NewRecorder()` |
| **Database (todos)** | | | | |
| Containers | testcontainers | multi-lang | https://testcontainers.com | PostgreSQL, MySQL, Redis, MongoDB reais |
| **E2E** | | | | |
| E2E | Playwright | 1.45+ | https://playwright.dev | Multi-browser, multi-language. Padrão |
| E2E alt | Cypress | 13.x | https://www.cypress.io | Para equipes já usando. Novo: usar Playwright |
| **Coverage** | | | | |
| Target | 80% de cobertura | — | — | Threshold obrigatório no CI. V8 para Node/TS |
| Branch coverage | `c8` / `istanbul` | — | — | Branch coverage > line coverage como indicador |
| **Performance** | | | | |
| Load testing | k6 | 0.51+ | https://k6.io | JavaScript. Integra com Grafana |
| Load testing alt | Artillery | 2.x | https://artillery.io | YAML config. Node/WebSocket |
| **Contract Testing** | | | | |
| Consumer-driven | Pact | 12.x (pact-js) | https://pact.io | Para microservices |

## Pirâmide de Testes

```
/\ E2E (Playwright) — 10%
/ \ Tempo: minutos
/----\
/ \ Integration — 20%
/ \ Tempo: segundos
/----------\
/ \ Unit — 70%
/ \ Tempo: milissegundos
/________________\

Regras:
- Unitários: sem I/O real, sem rede, sem banco
- Integração: banco real (testcontainers), sem E2E browser
- E2E: apenas happy paths e fluxos críticos de negócio
- Coverage: ≥ 80% statements (branch ≥ 70%). 100% em código de segurança/auth
- Separar unit de integração por build tag / marker / tag (nunca juntos no CI)
```

## Mutation, Contract e Load (cross-language)

| Categoria | Ferramenta | Versão | Notas |
|---|---|---|---|
| Mutation | Stryker | 8.x | JS/TS,.NET, Java. Score mínimo 60%, rodar weekly |
| Mutation (Java) | PITest | latest | Alternativa Java-only ao Stryker |
| Contract | Pact | 10.x+ | Consumer-driven, polyglot (microservices) |
| Load (Python) | Locust | 2.x | Python-native, distribuído |
| Load (geral) | k6 | 0.50+ | JS, Go-native output, integra Grafana/Prometheus |

> Versões acima são os pisos do ADR-021. Alvejar a última versão estável de cada ferramenta.
