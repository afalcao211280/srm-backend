# Runbook — SRM Credit Engine Backend

## Subir a stack completa

```bash
cp .env.example .env
docker compose up -d db
docker compose up -d api
docker compose --profile carga run --rm carga
```

## Portas e URLs

| Serviço | Porta | URL |
|---|---|---|
| Frontend | 3000 | http://localhost:3000 |
| Backend | 8080 | http://localhost:8080 |
| Backend healthz | 8080 | http://localhost:8080/healthz |
| Backend readyz | 8080 | http://localhost:8080/readyz |
| Backend metrics | 8080 | http://localhost:8080/metrics |
| API | 8080 | http://localhost:8080/api/v1/* |
| PostgreSQL | 5432 | postgres://srm:srm@localhost:5432/srm |

A porta do backend é publicada apenas para acesso direto a Swagger, métricas e health checks durante a avaliação. Em produção, o caminho funcional passa pela origem do frontend.

## Verificar saúde

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/readyz
```

## Acompanhar logs

```bash
docker compose logs -f api
docker compose logs -f db
```

## Reiniciar descartando os dados

```bash
docker compose down -v
```

Isso apaga os dados persistidos.

## Problemas comuns

- **Porta ocupada**: pare o processo que está usando a porta ou troque o mapeamento.
- **Banco indisponível**: `docker compose ps` mostra o status. `docker compose logs db` mostra o motivo.
- **Variável obrigatória ausente**: a aplicação falha imediatamente na inicialização indicando a variável faltante.
