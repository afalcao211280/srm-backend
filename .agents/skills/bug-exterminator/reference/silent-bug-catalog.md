# Catálogo de Bugs Silenciosos

Foco em Go (stack principal do usuário), com notas multi-linguagem. Para cada categoria: **sintoma**, **gatilho em produção** e **o que procurar**. Use como checklist contra as superfícies de risco mapeadas no recon.

> Regra de ouro da caça: um bug só conta se você consegue descrever a sequência concreta de eventos, no ambiente real, que o faz disparar.

---

## 1. Concorrência

### 1.1 Data race / estado compartilhado mutável
- **Sintoma**: dado corrompido intermitente, valor "impossível", pânico aleatório (`concurrent map writes`).
- **Gatilho**: múltiplas goroutines (ou múltiplas requests HTTP, que são goroutines) acessando o mesmo `map`/slice/struct/contador sem sincronização. Aparece sob carga/concorrência — invisível em teste single-thread.
- **Procurar**: variáveis de pacote/globais mutáveis; campos de struct compartilhada escritos sem `sync.Mutex`/`atomic`; `map` acessado de handlers HTTP; caches em memória; `sync.WaitGroup` reusado errado. **Rode `go test -race`.** Em Go, lembre que cada request roda em sua própria goroutine.

### 1.2 Goroutine leak
- **Sintoma**: memória e nº de goroutines sobem lentamente até OOMKill; throughput degrada.
- **Gatilho**: goroutine bloqueada para sempre num channel sem receiver/sender, ou esperando um `context` que nunca é cancelado. Cada request vaza uma → acumula em produção, não em teste curto.
- **Procurar**: `go func(){...}` que escreve em channel sem garantia de leitura; falta de `ctx` com cancelamento; `select` sem `case <-ctx.Done()`; channels sem buffer onde o caller pode desistir. Cheque `runtime.NumGoroutine()` em profiling.

### 1.3 Deadlock / lock mal usado
- **Sintoma**: requests travam, timeout em cascata, pod fica "vivo" mas não responde (health check passa, trabalho não anda).
- **Gatilho**: ordem inconsistente de aquisição de locks; lock mantido durante I/O lento; `RLock` que tenta virar `Lock`; esquecer `Unlock` em path de erro (use `defer`).
- **Procurar**: múltiplos mutexes; `Lock()` sem `defer Unlock()`; lock mantido em volta de chamada de rede/banco; `sync.Mutex` copiado (passar struct com mutex por valor copia o lock).

### 1.4 Loop variable capture (Go < 1.22) e closures
- **Sintoma**: todas as goroutines/closures veem o último valor do loop.
- **Gatilho**: `for _, v := range xs { go func(){ use(v) }() }` em Go < 1.22 (corrigido em 1.22+). Cheque a versão do `go.mod`.
- **Procurar**: closures dentro de loop capturando a variável de iteração; mesmo padrão em `defer` dentro de loop.

### 1.5 Estado em memória com múltiplas réplicas
- **Sintoma**: comportamento inconsistente entre requests (cache "some", rate limit não bate, sessão perdida).
- **Gatilho**: assumir um único processo, mas rodar com N réplicas atrás de load balancer. Cada pod tem sua própria memória.
- **Procurar**: cache/contador/rate-limiter/lock em memória local; "singleton" que deveria ser distribuído (Redis); afinidade de sessão assumida. **Bug clássico de quem testa local (1 processo) e roda em k8s (N pods).**

---

## 2. Manejo de erro

### 2.1 Erro engolido / ignorado
- **Sintoma**: falha silenciosa — o sistema "continua" com dado errado ou incompleto, e ninguém percebe até o efeito aparecer longe da causa.
- **Gatilho**: sempre que o caminho de erro é exercitado (dependência lenta/fora, input ruim) — o que mais acontece em produção.
- **Procurar**: `_ = f()` ou `f()` sem checar retorno de erro; `if err != nil { return nil }` (engole); `catch {}` vazio; `err` logado e fluxo continua como se tivesse dado certo; erro de `defer Close()` ignorado em escrita (pode esconder dado não-flushado).

### 2.2 Erro perdido no wrap / contexto apagado
- **Sintoma**: log diz "erro" mas não dá para saber a causa raiz; `errors.Is/As` falha.
- **Procurar**: `fmt.Errorf("...: %v", err)` em vez de `%w` (perde a cadeia); recriar erro genérico em vez de propagar; sentinela comparado com `==` quando foi feito wrap.

### 2.3 `nil` interface != `nil`
- **Sintoma**: `if err != nil` é verdadeiro mesmo "sem erro".
- **Gatilho**: retornar um ponteiro concreto nil dentro de uma interface (`error`). A interface carrega o tipo, então não é nil.
- **Procurar**: funções que retornam `*MyError` atribuído a `error`; `var err error = (*T)(nil)`.

### 2.4 Panic em produção
- **Sintoma**: a goroutine/processo morre; sem recover no boundary, derruba o pod.
- **Procurar**: index out of range, nil deref, type assertion sem `, ok`, divisão por zero, `panic()` em path alcançável por input externo. Confirme que o servidor tem recover no boundary (middleware) — e que o recover não engole erro de forma que esconde o problema.

---

## 3. Recursos e vazamentos

### 3.1 Recurso não fechado (conn, file, rows, body)
- **Sintoma**: esgotamento de pool de conexões, "too many open files", file descriptors vazando → erros sob carga.
- **Procurar**: `sql.Rows` sem `defer rows.Close()`; `http.Response.Body` não fechado (vaza conexão do pool e impede keep-alive); arquivo aberto sem `Close`; `defer` dentro de loop (acumula até o fim da função — fechar manualmente no loop). Em outras linguagens: `with`/`try-with-resources` ausente.

### 3.2 Context não propagado / sem timeout
- **Sintoma**: chamada que deveria desistir fica pendurada; quando o downstream trava, todo o sistema trava junto (thread/goroutine/conn presa).
- **Gatilho**: dependência lenta ou que parou de responder — sempre acontece eventualmente.
- **Procurar**: `context.Background()`/`context.TODO()` em chamada de rede/banco que deveria herdar o ctx do request; HTTP client sem `Timeout`; query sem `QueryContext`; falta de `ctx` no caminho até a borda; deadline do request não respeitada.

### 3.3 Crescimento ilimitado de memória
- **Sintoma**: memória sobe até o limite do container → **OOMKill**, pod reinicia.
- **Gatilho**: carga real e dados reais (maiores que o fixture de teste).
- **Procurar**: ler corpo/arquivo inteiro na memória (`io.ReadAll` sem limite — use `io.LimitReader`); slice/map/cache que só cresce e nunca é podado; acumular resultados de query sem paginação; buffers que crescem com input externo. Lembre do limite de memória do container.

---

## 4. Valores nulos, zero e de borda

### 4.1 Nil pointer / nil map
- **Procurar**: deref de ponteiro que pode ser nil (retorno de função, campo opcional, resultado de `map[k]` ausente); **escrita em `map` nil** (`var m map[k]v; m[x]=y` → panic); slice nil indexada.

### 4.2 Zero value tratado como "ausente"
- **Sintoma**: `0`, `""`, `false`, `time.Time{}` zero confundidos com "não informado".
- **Gatilho**: input legítimo que por acaso é o zero value (quantidade 0, string vazia válida, ano 1).
- **Procurar**: lógica que usa zero value para sinalizar ausência; JSON sem ponteiro/`omitempty` mal usado; `if x != 0` quando 0 é valor válido.

### 4.3 Borda de coleção
- **Procurar**: lista vazia (divisão por len, acesso a `[0]`); lista enorme (N+1, memória, timeout); off-by-one em fatiamento (`s[i:j]`); paginação com último página parcial; índice calculado que pode estourar.

---

## 5. Tempo e número

### 5.1 Timezone / data
- **Sintoma**: "funciona local, erra em produção" — clássico de fuso. Datas off-by-one-day, agendamento na hora errada, comparação de dia errada.
- **Gatilho**: servidor em UTC, usuário/negócio em America/Sao_Paulo; horário de verão; `time.Now()` sem location explícita.
- **Procurar**: `time.Now()` comparado com data de negócio sem normalizar fuso; parsing de data sem timezone; truncar para "dia" em UTC; cron expressando hora local. Pergunte qual o fuso de produção.

### 5.2 Overflow / truncamento / float
- **Procurar**: soma/multiplicação de inteiros que pode estourar `int32`; conversão `int64→int32`/`int→int8` truncando; dinheiro em `float` (use inteiro de centavos ou decimal); comparação de float com `==`; `len()` de algo enorme.

---

## 6. Banco de dados

### 6.1 N+1
- **Sintoma**: latência aceitável com poucos registros, timeout/lentidão quando a lista cresce em produção.
- **Procurar**: query dentro de loop sobre resultados de outra query; ORM carregando relações lazy em iteração; falta de `JOIN`/`IN`/batch.

### 6.2 Transação ausente / mal delimitada
- **Sintoma**: estado parcial após falha (debitou e não creditou); dado inconsistente.
- **Procurar**: múltiplos writes que deveriam ser atômicos sem `BEGIN/COMMIT`; commit sem checar erro; rollback ausente no path de erro (`defer tx.Rollback()`); transação mantida aberta durante I/O lento (segura conexão e locks).

### 6.3 Pool e isolamento
- **Procurar**: pool de conexões pequeno demais para a concorrência (esgota sob carga) ou grande demais (estoura o `max_connections` do Postgres com N réplicas); nível de isolamento assumido (lost update sob concorrência); `SELECT ... FOR UPDATE` ausente onde há read-modify-write concorrente; falta de índice → full scan que só dói com a tabela cheia em prod.

---

## 7. Rede / HTTP / integrações

### 7.1 Falta de timeout e retry sem backoff
- **Sintoma**: quando um downstream degrada, o problema se espalha; retries amplificam a carga e **derrubam o que estava só lento** (retry storm / thundering herd).
- **Procurar**: `http.Client{}` sem `Timeout`; retry em loop sem backoff exponencial + jitter; retry em erro não-idempotente; ausência de circuit breaker; sem limite de tentativas.

### 7.2 Idempotência
- **Sintoma**: efeito duplicado (cobrança, e-mail, mensagem) quando há retry, redelivery de fila, ou clique duplo.
- **Gatilho**: retries (que você adicionou no item anterior), at-least-once de mensageria, timeout no client mas sucesso no server.
- **Procurar**: operação com efeito colateral sem chave de idempotência; consumer de fila que assume exactly-once; POST que cria sem dedup.

### 7.3 TOCTOU (time-of-check to time-of-use)
- **Sintoma**: "verifiquei que existe/tinha saldo, então usei" — e entre o check e o use, mudou.
- **Procurar**: `if exists { create }`, `if saldo >= x { debita }` sem lock/transação/constraint atômica; checagem de arquivo seguida de uso.

---

## 8. Serialização e contrato

- **Procurar**: `omitempty` que apaga campo legítimo com valor zero; número grande em JSON virando float e perdendo precisão (IDs int64 em JS/JSON); campo opcional desserializado como zero value silenciosamente; mudança de contrato que quebra consumidor antigo; enum recebendo valor desconhecido tratado como zero/default.

---

## 9. Lifecycle do ambiente

- **Graceful shutdown ausente**: pod recebe SIGTERM (deploy, scale-down), mata requests em andamento, perde mensagens em processamento. Procurar tratamento de sinal + drenagem de conexões/consumers.
- **Startup/readiness**: começa a receber tráfego antes de dependências estarem prontas (readiness probe incorreta); cache frio causando timeout no primeiro hit.
- **Config divergente dev/prod**: default no código diferente do valor de produção; env var ausente caindo em default perigoso; feature flag; `DEBUG=true` vazando. Procurar tudo que lê env/config e o que acontece se faltar.
- **Idempotência de migração/job**: job/cron que roda em N réplicas simultaneamente sem lock; migração não-reentrante.

---

## Notas multi-linguagem (quando não for Go)
- **Python**: argumento default mutável (`def f(x=[])`); `except:` nu engolindo tudo; GIL não protege operações compostas; recursos sem `with`; `datetime` naive vs aware.
- **Java**: `NullPointerException` em autoboxing; `equals`/`hashCode` inconsistentes; recursos sem try-with-resources; `ExecutorService` não encerrado; `SimpleDateFormat` não thread-safe.
- **TypeScript/Node**: promise sem `await`/`.catch` (unhandled rejection); `==` vs `===`; `undefined` vs `null`; float para dinheiro; event loop bloqueado por trabalho síncrono; `JSON.parse` de input externo sem validação.
