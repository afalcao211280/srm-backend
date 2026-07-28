# ADR-013 — Next.js 16 e TypeScript 5.9.x

**Decisão:** `next` 16.2.12, `react` 19.2.8, `typescript` **5.9.3** (fixado).

**Armadilha:** `typescript@latest` resolve para `7.0.2`, que **não pode ser usado**. O pacote 7.0 ships compilador nativo em Go e removeu `lib/typescript.js`; o Next.js 16 carrega o pacote como módulo JS e reporta não instalado. `typescript-eslint@8.65.0` declara peer `typescript >=4.8.4 <6.1.0` e encerrou o pedido de suporte a TS 7 como *not planned*. Fixamos a linha 5.9.x.
