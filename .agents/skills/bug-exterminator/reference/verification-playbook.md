# Playbook de Verificação — Refutar antes de Reportar

Um falso positivo custa confiança e gera retrabalho. Antes de chamar algo de bug, tente **derrubá-lo**. Só sobrevive o que você não conseguiu refutar. Este playbook calibra confiança e evita relatórios cheios de "talvez".

## A mentalidade: caçador cético

Você não está procurando confirmar seus medos — está tentando provar que cada candidato **não** é um bug. O que resistir a essa tentativa de refutação é real.

Para cada candidato, responda nesta ordem:

### 1. Descreva a sequência concreta que dispara o bug
Escreva a cadeia de eventos, com valores reais, no ambiente real:
> "Request A entra, adquire X; antes de liberar, request B entra e lê X parcial; A falha em `f()` e retorna sem reverter; B vê estado inconsistente e grava Y errado no banco."

**Se você não consegue escrever essa cadeia, provavelmente não é um bug** — é um desconforto teórico. Marque como suspeito ou descarte.

### 2. Procure a proteção que já pode existir
O bug pode estar coberto rio acima ou rio abaixo. Antes de reportar, verifique:
- Há um lock/transação/constraint que serializa o acesso?
- O caller já valida o input / já trata o erro / já faz retry?
- Existe uma constraint de banco (UNIQUE, NOT NULL, FK) que impede o estado ruim?
- Há um middleware (recover, timeout, auth) que intercepta?
- O `context` já carrega um deadline herdado?
- A versão da linguagem já corrige isso (ex.: Go 1.22 e loop capture)?

Leia o entorno — não só a linha. Muito "bug" some quando você lê a função inteira e o caller.

### 3. Reproduza quando viável
A prova mais forte é um teste que falha:
- **Concorrência**: escreva um teste que roda N goroutines e rode com `go test -race`. O race detector é juiz objetivo. Sem ele, race é especulação.
- **Input/borda**: um teste com o input específico (lista vazia, payload gigante, número negativo, data no fuso problemático) que produz o resultado errado.
- **Leak**: um loop que exercita o caminho e mede `runtime.NumGoroutine()` ou memória antes/depois.
- **N+1 / performance**: conte queries (log do driver) com lista pequena vs grande; ou benchmark.
- **Lógica**: o caso que viola a invariante, demonstrado.

Se não dá para reproduzir automatizado, descreva o experimento manual que confirmaria (e diga que não foi executado).

### 4. Classifique a confiança — e seja honesto
- **Confirmado**: reproduzi (teste/`-race`/repro) OU o raciocínio é incontestável e a proteção não existe. Reporte como bug.
- **Suspeito**: plausível, mas não consegui reproduzir nem garantir que não há proteção. Reporte como suspeito, dizendo exatamente **o que falta** para confirmar (ex.: "preciso saber se roda com >1 réplica", "preciso confirmar o fuso do banco").
- **Refutado**: encontrei a proteção / não consigo montar o cenário / a versão já corrige. **Descarte** (não polua o relatório). Se foi um quase-bug instrutivo, pode mencionar de passagem.

## Verificação em camadas para achados de alto risco

Para um bug **Crítico/Alto**, não confie num único ângulo. Confirme por mais de uma lente antes de afirmar:
- **Correção**: a lógica está mesmo errada?
- **Alcançabilidade**: o caminho é atingível por input/evento real em produção (não código morto)?
- **Ambiente**: a condição de produção que dispara existe de fato (a concorrência acontece? roda com N réplicas? o fuso diverge?)?

Um achado que passa nas três é sólido. Se falha em "alcançabilidade" ou "ambiente", rebaixe a severidade ou mova para suspeito.

## Antes de propor a correção

- A correção resolve a **causa**, não o sintoma? (adicionar retry não conserta a race; conserta a sincronização).
- A correção respeita as `.cursor/rules` e os padrões do projeto?
- A correção não introduz um bug novo? (ex.: adicionar lock pode criar deadlock; adicionar timeout pode cortar operação legítima longa).
- É a mudança **mínima** que resolve? Evite refatorar de carona.

## Depois de aplicar (sob aprovação)

1. Rode os testes — com `-race` se foi concorrência.
2. Rode o lint (`golangci-lint run` ou equivalente).
3. Confirme que o cenário do bug **não dispara mais** (idealmente o teste que falhava agora passa).
4. Cheque que nada quebrou em volta.
5. Reporte: o que era, por que disparava, o que mudou, e como você verificou.

## Formato do relatório (por bug)

```
[SEVERIDADE] Título curto — arquivo:linha
Impacto em produção: <o que o sistema/usuário sente quando dispara>
Gatilho: <a condição real de produção que ativa>
Evidência: <trecho + cenário + teste/repro, se houver>
Confiança: confirmado | suspeito (falta: ___)
Correção proposta: <mudança mínima; aguardando aprovação para aplicar>
```

Ordene o relatório por risco (severidade × probabilidade), não pela ordem dos arquivos. Abra com um resumo de 1-2 linhas: quantos confirmados, quantos suspeitos, e o de maior risco.
