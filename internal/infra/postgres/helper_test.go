//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/srm-asset/srm-backend/internal/platform/migracao"
)

// testPool é compartilhado entre todos os testes deste pacote: subir um
// container por teste seria lento e desnecessário, já que cada teste faz
// truncate() antes de rodar. O container real (não mock) valida constraint,
// índice, tipo e transação de verdade — mockar o banco esconderia
// exatamente os bugs que o pacote quer pegar.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("srm_test"),
		postgres.WithUsername("srm"),
		postgres.WithPassword("srm"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "subir container de teste:", err)
		os.Exit(1)
	}
	defer func() { _ = container.Terminate(ctx) }()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection string:", err)
		os.Exit(1)
	}

	if err := migracao.Aplicar(dsn, "../../../migrations"); err != nil {
		fmt.Fprintln(os.Stderr, "aplicar migrations no container de teste:", err)
		os.Exit(1)
	}

	pool, err := NewPool(ctx, Config{DatabaseURL: dsn, MaxConns: 10, MinConns: 1, ConnTimeout: 5 * time.Second})
	if err != nil {
		fmt.Fprintln(os.Stderr, "criar pool de teste:", err)
		os.Exit(1)
	}
	defer pool.Close()
	testPool = pool

	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if testPool == nil {
		t.Fatal("pool de teste não inicializado — TestMain falhou?")
	}
	return testPool
}

func truncate(t *testing.T, pool *pgxpool.Pool) {
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
		if _, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			t.Fatalf("truncar %s: %v", table, err)
		}
	}
}
