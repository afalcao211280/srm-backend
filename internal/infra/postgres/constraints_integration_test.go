//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/srm-asset/srm-backend/internal/domain/money"
)

func sqlState(t *testing.T, err error) string {
	t.Helper()
	var pgErr *pgconn.PgError
	if err == nil {
		t.Fatal("esperado erro do postgres, obtido nil")
	}
	if !asPgError(err, &pgErr) {
		t.Fatalf("erro não é *pgconn.PgError: %T %v", err, err)
	}
	return pgErr.Code
}

func asPgError(err error, target **pgconn.PgError) bool {
	for err != nil {
		if pe, ok := err.(*pgconn.PgError); ok {
			*target = pe
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}

func TestConstraintChaveEstrangeiraInvalida(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	tx := transacaoFixture("66666666-6666-6666-6666-666666666661", f)
	tx.CedenteID = 999999 // cedente inexistente
	err := NewTransacaoRepo(pool).Criar(ctx, tx)
	if got := sqlState(t, err); got != pgerrcode.ForeignKeyViolation {
		t.Fatalf("esperado violação de FK (%s), obtido %s: %v", pgerrcode.ForeignKeyViolation, got, err)
	}
}

func TestConstraintValorFaceNaoPositivo(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	tx := transacaoFixture("66666666-6666-6666-6666-666666666662", f)
	tx.ValorFace = money.New(0)
	err := NewTransacaoRepo(pool).Criar(ctx, tx)
	if got := sqlState(t, err); got != pgerrcode.CheckViolation {
		t.Fatalf("esperado violação de check (%s), obtido %s: %v", pgerrcode.CheckViolation, got, err)
	}
}

func TestConstraintVencimentoNaoPosteriorOperacao(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	tx := transacaoFixture("66666666-6666-6666-6666-666666666663", f)
	tx.DataVencimento = tx.DataOperacao // igual, não posterior
	err := NewTransacaoRepo(pool).Criar(ctx, tx)
	if got := sqlState(t, err); got != pgerrcode.CheckViolation {
		t.Fatalf("esperado violação de check (%s), obtido %s: %v", pgerrcode.CheckViolation, got, err)
	}
}

func TestConstraintStatusForaDoDominio(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()
	f := criarFixture(t, ctx, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO transacoes (
			id, cedente_id, tipo_recebivel_id, moeda_titulo_id, moeda_pagamento_id,
			valor_face, valor_presente, valor_liquido, desagio,
			spread_aplicado, taxa_base_aplicada, data_operacao, data_vencimento, status, versao
		) VALUES ($1,$2,$3,$4,$5, 1000,950,950,50, 0.015,0.01, $6,$7,'INVALIDO',1)`,
		"66666666-6666-6666-6666-666666666664", f.CedenteID, f.TipoID, f.MoedaID, f.MoedaID,
		time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
	)
	if got := sqlState(t, err); got != pgerrcode.CheckViolation && got != pgerrcode.InvalidTextRepresentation {
		t.Fatalf("esperado violação de check/enum para status inválido, obtido %s: %v", got, err)
	}
}

func TestConstraintCodigoMoedaUnico(t *testing.T) {
	pool := setupTestDB(t)
	truncate(t, pool)
	ctx := context.Background()

	_, err := pool.Exec(ctx, `INSERT INTO moedas (codigo, nome, casas_decimais) VALUES ('BRL', 'Duplicado', 2)`)
	if got := sqlState(t, err); got != pgerrcode.UniqueViolation {
		t.Fatalf("esperado violação de unicidade (%s), obtido %s: %v", pgerrcode.UniqueViolation, got, err)
	}
}
