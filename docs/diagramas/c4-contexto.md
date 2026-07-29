# C4 Nível 1 — Contexto

```
[Operador]
    |
    | navega na UI / submete requisições
    v
[SRM Credit Engine]
    |
    |-- consome cotações --> [Provedor de Câmbio Externo]
    |
    |-- persiste em --------> [PostgreSQL]
```

Atores:
- **Operador**: pessoa da mesa de operação que simula e liquida recebíveis.
- **Provedor de Câmbio Externo**: provedor mockado de cotações (substituível em teste).
- **PostgreSQL**: banco de dados relacional com a precisão NUMERIC.

Sistema:
- **SRM Credit Engine**: backend Go + frontend Next.js operando como um único sistema de cessão de crédito.
