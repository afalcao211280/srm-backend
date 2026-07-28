// Package testdb centraliza a subida de um container PostgreSQL real
// (testcontainers-go) com as migrations aplicadas, para reuso pelos
// pacotes que precisam de testes de integração (postgres, report,
// handler). Container real, não mock — constraint, índice, tipo e
// transação só se provam contra o banco de verdade.
package testdb

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	infrapostgres "github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/migracao"
)

// Subir inicia um container Postgres, aplica as migrations e devolve um
// pool pronto para uso. Chamar a partir de TestMain; o cleanup encerra o
// container e fecha o pool.
func Subir(ctx context.Context) (*pgxpool.Pool, func(), error) {
	dsn, cleanupContainer, err := subirContainer(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := migracao.Aplicar(dsn, migrationsPath()); err != nil {
		cleanupContainer()
		return nil, nil, fmt.Errorf("aplicar migrations: %w", err)
	}
	pool, err := infrapostgres.NewPool(ctx, infrapostgres.Config{
		DatabaseURL: dsn, MaxConns: 10, MinConns: 1, ConnTimeout: 5 * time.Second,
	})
	if err != nil {
		cleanupContainer()
		return nil, nil, fmt.Errorf("criar pool de teste: %w", err)
	}
	cleanup := func() {
		pool.Close()
		cleanupContainer()
	}
	return pool, cleanup, nil
}

func subirContainer(ctx context.Context) (dsn string, cleanup func(), err error) {
	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("srm_test"),
		tcpostgres.WithUsername("srm"),
		tcpostgres.WithPassword("srm"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("subir container de teste: %w", err)
	}
	cleanup = func() { _ = container.Terminate(ctx) }
	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("connection string: %w", err)
	}
	return dsn, cleanup, nil
}

// migrationsPath resolve o diretório migrations/ a partir do caminho
// absoluto deste arquivo fonte (runtime.Caller), funcionando independente
// do diretório de trabalho com o qual `go test` roda cada pacote chamador.
func migrationsPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// este arquivo vive em internal/platform/testdb/testdb.go — a raiz do
	// módulo fica três níveis acima.
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(thisFile))))
	return filepath.Join(root, "migrations")
}

func Truncate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tables := []string{
		"transacao_auditoria",
		"transacoes",
		"taxas_base",
		"cotacoes_cambio",
		"cedentes",
	}
	for _, table := range tables {
		if _, err := pool.Exec(ctx, "DELETE FROM "+table); err != nil {
			t.Fatalf("truncar %s: %v", table, err)
		}
	}
}
