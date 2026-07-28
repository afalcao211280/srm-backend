# Recon do Projeto e do Ambiente

Antes de caçar qualquer bug, construa o mapa. Bug silencioso é função do ambiente: o mesmo código é correto num lugar e incidente em outro. Este checklist evita auditar no vácuo.

> Inferir do código/config primeiro; perguntar ao usuário só o que não der para descobrir. Mas **não assuma** — uma suposição errada sobre o ambiente gera diagnóstico que erra o alvo.

## 1. O que é o projeto

- **Linguagem e versão exata** — leia `go.mod` (`go 1.22`?), `.python-version`, `package.json` engines, `pom.xml`. A versão muda o que é bug (ex.: loop capture em Go < 1.22).
- **O que ele faz** — API HTTP? worker de fila? cron/batch? CLI? gRPC? Isso define o modelo de concorrência e os modos de falha.
- **Caminhos críticos** — o que dói se quebrar? (pagamento, provisionamento, billing, auth). Priorize a caça aqui.
- **Dependências (libs)** — frameworks web, ORM/driver de banco, client HTTP, cliente de fila. Leia o manifesto de deps.
- **Rules do projeto** — `.cursor/rules/*.mdc`, CONTRIBUTING, ADRs. Definem invariantes e convenções que, se violadas, são bug.

## 2. Onde e como roda (o ambiente)

### Container e recursos
- É containerizado? Leia o `Dockerfile` — roda como root? multi-stage? que processo?
- **Limites de CPU/memória** — k8s `resources.limits/requests`, `docker run --memory`. **Memory limit é teto duro: estourar = OOMKill.** Isso transforma "uso um pouco de memória a mais" em incidente.
- Tem profiling/métricas de memória e goroutines?

### Escala e topologia
- **Quantas réplicas?** (Deployment `replicas`, HPA min/max). Se > 1, **estado em memória não é compartilhado** — cache/contador/rate-limiter local viram bug.
- Como escala? HPA por CPU/memória/custom? Carga é estável ou tem picos?
- Há load balancer / afinidade de sessão? Stickiness é assumida?

### Estado e dados
- Onde o estado vive? Memória do pod (efêmero!), banco, cache distribuído (Redis), object storage?
- Banco: qual, qual `max_connections`, tamanho do pool por réplica (pool × réplicas ≤ max_connections?), índices, volume de dados real vs dev.
- Mensageria: at-least-once? redelivery? ordem garantida? DLQ?

### Rede e dependências externas
- Quais serviços externos são chamados? Para **cada um**: o que acontece se ficar lento? se cair? tem timeout configurado? retry? circuit breaker?
- Latência interna esperada; nginx/proxy no caminho (timeouts do proxy vs da app).

### Tempo, locale, config
- **Timezone de produção** vs dev (servidor UTC? negócio America/Sao_Paulo? horário de verão?). Fonte nº 1 de "funciona local, erra em prod".
- Locale/encoding (formatação de número, collation de banco).
- Como a config chega? env vars, ConfigMap, Secret, arquivo. O que acontece se uma faltar — cai em default perigoso? Há divergência conhecida entre dev e prod?

### Lifecycle
- Graceful shutdown: trata SIGTERM? drena requests/consumers? (pods morrem o tempo todo: deploy, scale-down, rescheduling).
- Readiness/liveness probes: só recebe tráfego quando pronto? probe mede a coisa certa?
- Jobs/migrações: rodam em quantas réplicas ao mesmo tempo? são reentrantes/idempotentes?

## 3. Os contratos de robustez

Cheque a presença e os valores de:
- Timeouts (HTTP client, server read/write, query, context deadline).
- Limites (tamanho de payload/body, paginação, rate limit, tamanho de pool).
- Políticas de retry (backoff? jitter? só idempotente? limite?).
- Idempotência em operações com efeito colateral.

## 4. Perguntas para o usuário (quando não der para inferir)

Faça só as que importam para o que você vai auditar:
- "Quantas réplicas isso roda em produção e qual o limite de memória do container?"
- "Qual o timezone de produção e o do banco?"
- "Esse código assume processo único ou já roda com múltiplas réplicas?"
- "Qual o perfil de carga/concorrência esperado nesse endpoint?"
- "O que é crítico aqui — o que não pode falhar de jeito nenhum?"
- "Quais dependências externas isso chama e o que já aconteceu quando elas caíram?"

## Saída do recon
Antes de ir para a Fase 2, você deve conseguir completar:
> "Este é um **[tipo de serviço]** em **[linguagem/versão]** que roda em **[ambiente: N réplicas, limite X de memória]**, com estado em **[onde]**, dependendo de **[externos]**, em fuso **[tz]**. O que mais dói se quebrar é **[caminho crítico]**. Os modos de falha mais prováveis do ambiente são **[concorrência/carga/dep lenta/fuso/...]**."

Se não consegue preencher isso, faça mais recon antes de auditar.
