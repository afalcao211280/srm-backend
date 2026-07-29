package recebivel

import (
	"errors"
	"testing"
)

func TestStrategyDuplicata(t *testing.T) {
	r := DefaultRegistry()
	s, err := r.Resolve("DUPLICATA_MERCANTIL")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if s.Spread() != "0.015" {
		t.Fatalf("spread duplicata: esperado 0.015, obtido %s", s.Spread())
	}
}

func TestStrategyCheque(t *testing.T) {
	r := DefaultRegistry()
	s, err := r.Resolve("CHEQUE_PRE_DATADO")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if s.Spread() != "0.025" {
		t.Fatalf("spread cheque: esperado 0.025, obtido %s", s.Spread())
	}
}

func TestStrategyInexistente(t *testing.T) {
	r := DefaultRegistry()
	_, err := r.Resolve("INEXISTENTE")
	if err == nil {
		t.Fatalf("esperado erro para tipo sem strategy")
	}
	var alvo *TipoSemStrategyError
	if !errors.As(err, &alvo) {
		t.Fatalf("erro: tipo inesperado %T: %v", err, err)
	}
}
