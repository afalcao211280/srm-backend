package precificacao

import (
	"testing"
	"time"

	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
)

func data(d string) time.Time {
	t, _ := time.Parse("2006-01-02", d)
	return t
}

func novaEntrada(valorFace, taxaBase string, op, venc string, tipo string, comCot bool, cot string, moedaPag string) Entrada {
	vf, _ := money.FromString(valorFace)
	tb, _ := money.FromString(taxaBase)
	ct, _ := money.FromString(cot)
	return Entrada{
		ValorFace:      vf,
		DataOperacao:   data(op),
		DataVencimento: data(venc),
		TipoRecebivel:  tipo,
		TaxaBase:       tb,
		TemCotacao:     comCot,
		Cotacao:        ct,
		MoedaTitulo:    "BRL",
		MoedaPagamento: moedaPag,
	}
}

func TestDuplicata45Dias(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	wantPrazo := money.MustFromString("1.5")
	if r.PrazoQuantizado().Cmp(wantPrazo) != 0 {
		t.Fatalf("prazo: esperado 1.5, obtido %s", r.PrazoQuantizado())
	}
	if r.ValorLiquido.String() != "9636.39" {
		t.Fatalf("valor líquido: esperado 9636.39, obtido %s", r.ValorLiquido)
	}
	if r.ValorPresente8Casas.String() != "9636.38630878" {
		t.Fatalf("valor presente 8: esperado 9636.38630878, obtido %s", r.ValorPresente8Casas)
	}
	if r.Desagio.String() != "363.61" {
		t.Fatalf("deságio: esperado 363.61, obtido %s", r.Desagio)
	}
}

func TestDuplicata30Dias(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-07-31", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	wantPrazo := money.MustFromString("1")
	if r.PrazoQuantizado().Cmp(wantPrazo) != 0 {
		t.Fatalf("prazo: esperado 1, obtido %s", r.PrazoQuantizado())
	}
	if r.ValorLiquido.String() != "9756.10" {
		t.Fatalf("valor líquido: esperado 9756.10, obtido %s", r.ValorLiquido)
	}
}

func TestCheque60Dias(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-30", "CHEQUE_PRE_DATADO", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	if r.ValorLiquido.String() != "9335.11" {
		t.Fatalf("valor líquido: esperado 9335.11, obtido %s", r.ValorLiquido)
	}
}

func TestCheque180DiasAltaTaxa(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("250000.00", "0.0125", "2026-07-01", "2026-12-28", "CHEQUE_PRE_DATADO", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	if r.ValorLiquido.String() != "200452.45" {
		t.Fatalf("valor líquido: esperado 200452.45, obtido %s", r.ValorLiquido)
	}
}

func TestDuplicata1Dia(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("1000.00", "0.01", "2026-07-01", "2026-07-02", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	if r.ValorLiquido.String() != "999.18" {
		t.Fatalf("valor líquido: esperado 999.18, obtido %s", r.ValorLiquido)
	}
}

func TestCrossCurrencyBRLparaUSD(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", true, "5.4321", "USD")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	if r.ValorLiquido.String() != "1773.97" {
		t.Fatalf("valor líquido USD: esperado 1773.97, obtido %s", r.ValorLiquido)
	}
	if r.ValorPresente8Casas.String() != "9636.38630878" {
		t.Fatalf("valor presente BRL: esperado 9636.38630878, obtido %s", r.ValorPresente8Casas)
	}
}

func TestBaseNaoPositiva(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "-1.5", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	_, err := m.Precificar(e)
	if err == nil {
		t.Fatalf("esperado erro de base não positiva")
	}
	if _, ok := err.(*BaseNaoPositivaError); !ok {
		t.Fatalf("erro: tipo inesperado %T: %v", err, err)
	}
}

func TestVencimentoIgualOperacao(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-07-01", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	_, err := m.Precificar(e)
	if err == nil {
		t.Fatalf("esperado erro de prazo inválido")
	}
}

func TestVencimentoAnteriorOperacao(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-15", "2026-07-01", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	_, err := m.Precificar(e)
	if err == nil {
		t.Fatalf("esperado erro de prazo inválido")
	}
}

func TestTipoSemStrategy(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "INEXISTENTE", false, "1", "BRL")
	_, err := m.Precificar(e)
	if err == nil {
		t.Fatalf("esperado erro de strategy ausente")
	}
}

func TestCotacaoInexistenteCrossCurrency(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", false, "1", "USD")
	_, err := m.Precificar(e)
	if err == nil {
		t.Fatalf("esperado erro de cotação ausente em cross-currency")
	}
}

func TestReprodutibilidade(t *testing.T) {
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	r1, _ := m.Precificar(e)
	r2, _ := m.Precificar(e)
	if r1.ValorPresente.Cmp(r2.ValorPresente) != 0 {
		t.Fatalf("valor presente não é reproduzível: %s vs %s", r1.ValorPresente, r2.ValorPresente)
	}
	if r1.ValorLiquido.Cmp(r2.ValorLiquido) != 0 {
		t.Fatalf("valor líquido não é reproduzível: %s vs %s", r1.ValorLiquido, r2.ValorLiquido)
	}
}

func TestFusoHorarioResultadoIgual(t *testing.T) {
	// A data de operação é fixa; o fuso do container não entra no cálculo.
	m := NewMotor(recebivel.DefaultRegistry())
	e := novaEntrada("10000.00", "0.01", "2026-07-01", "2026-08-15", "DUPLICATA_MERCANTIL", false, "1", "BRL")
	r, err := m.Precificar(e)
	if err != nil {
		t.Fatalf("precificar: %v", err)
	}
	if r.ValorLiquido.String() != "9636.39" {
		t.Fatalf("valor líquido sob fuso: esperado 9636.39, obtido %s", r.ValorLiquido)
	}
}
