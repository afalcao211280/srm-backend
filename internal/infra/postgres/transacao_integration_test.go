//go:build integration

package postgres

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/srm-asset/srm-backend/internal/domain/money"
)

func TestLiquidacaoConcorrencia(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	truncate(t, pool)
	ctx := context.Background()

	repo := NewTransacaoRepo(pool)
	id := "11111111-1111-1111-1111-111111111111"
	vf, _ := money.FromString("1000.00")
	vp, _ := money.FromString("950.00")
	vl, _ := money.FromString("950.00")
	des, _ := money.FromString("50.00")
	sp, _ := money.FromString("0.015")
	tb, _ := money.FromString("0.01")
	op := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	vc := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	tx := Transacao{
		ID:               id,
		CedenteID:        1,
		TipoRecebivelID:  1,
		MoedaTituloID:    1,
		MoedaPagamentoID: 1,
		ValorFace:        vf,
		ValorPresente:    vp,
		ValorLiquido:     vl,
		Desagio:          des,
		SpreadAplicado:   sp,
		TaxaBaseAplicada: tb,
		DataOperacao:     op,
		DataVencimento:   vc,
		Status:           StatusPendente,
	}
	if err := repo.Criar(ctx, tx); err != nil {
		t.Fatalf("criar: %v", err)
	}
	const N = 10
	var sucessos, conflitos int32
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, err := repo.Liquidar(ctx, id, 1, time.Now().UTC())
			if err == nil {
				atomic.AddInt32(&sucessos, 1)
			} else if err == ErroConflitoLiquidacao {
				atomic.AddInt32(&conflitos, 1)
			}
		}()
	}
	wg.Wait()
	if sucessos != 1 {
		t.Fatalf("esperado 1 sucesso, obtido %d", sucessos)
	}
	if conflitos != N-1 {
		t.Fatalf("esperado %d conflitos, obtido %d", N-1, conflitos)
	}
	atual, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if atual.Versao != 2 {
		t.Fatalf("versão: esperado 2, obtido %d", atual.Versao)
	}
}

func TestLiquidacaoVersaoDesatualizada(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	truncate(t, pool)
	ctx := context.Background()

	repo := NewTransacaoRepo(pool)
	id := "22222222-2222-2222-2222-222222222222"
	vf, _ := money.FromString("1000.00")
	vp, _ := money.FromString("950.00")
	vl, _ := money.FromString("950.00")
	des, _ := money.FromString("50.00")
	sp, _ := money.FromString("0.015")
	tb, _ := money.FromString("0.01")
	if err := repo.Criar(ctx, Transacao{
		ID: id, CedenteID: 1, TipoRecebivelID: 1,
		MoedaTituloID: 1, MoedaPagamentoID: 1,
		ValorFace: vf, ValorPresente: vp, ValorLiquido: vl, Desagio: des,
		SpreadAplicado: sp, TaxaBaseAplicada: tb,
		DataOperacao: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		DataVencimento: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Status: StatusPendente,
	}); err != nil {
		t.Fatalf("criar: %v", err)
	}
	if _, err := repo.Liquidar(ctx, id, 1, time.Now().UTC()); err != nil {
		t.Fatalf("primeira liquidação: %v", err)
	}
	_, err := repo.Liquidar(ctx, id, 1, time.Now().UTC())
	if err != ErroConflitoLiquidacao {
		t.Fatalf("segunda liquidação com versão antiga: esperado conflito, obtido %v", err)
	}
}

func TestRoundTripDecimal(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()
	truncate(t, pool)
	ctx := context.Background()

	repo := NewTransacaoRepo(pool)
	id := "33333333-3333-3333-3333-333333333333"
	vpOriginal, _ := money.FromString("9636.38630878")
	vf, _ := money.FromString("10000.00")
	vl, _ := money.FromString("9636.39")
	des, _ := money.FromString("363.61")
	sp, _ := money.FromString("0.015")
	tb, _ := money.FromString("0.01")
	if err := repo.Criar(ctx, Transacao{
		ID: id, CedenteID: 1, TipoRecebivelID: 1,
		MoedaTituloID: 1, MoedaPagamentoID: 1,
		ValorFace: vf, ValorPresente: vpOriginal, ValorLiquido: vl, Desagio: des,
		SpreadAplicado: sp, TaxaBaseAplicada: tb,
		DataOperacao: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		DataVencimento: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		Status: StatusPendente,
	}); err != nil {
		t.Fatalf("criar: %v", err)
	}
	lida, err := repo.PorID(ctx, id)
	if err != nil {
		t.Fatalf("por id: %v", err)
	}
	if lida.ValorPresente.Cmp(vpOriginal) != 0 {
		t.Fatalf("round-trip: esperado %s, obtido %s", vpOriginal, lida.ValorPresente)
	}
}
