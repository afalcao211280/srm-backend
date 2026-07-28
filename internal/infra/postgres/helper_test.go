//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://srm:srm@localhost:5432/srm_test?sslmode=disable"
	}
	pool, err := NewPool(context.Background(), Config{
		DatabaseURL: dsn,
		MaxConns:    5,
		ConnTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Skipf("banco de teste indisponível: %v", err)
	}
	return pool
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
