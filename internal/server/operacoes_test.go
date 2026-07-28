package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/handler"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

// Estes testes cobrem apenas os caminhos que retornam ANTES de tocar
// qualquer repositório (JSON malformado, campo obrigatório ausente,
// decimal inválido) — por isso os repositórios são construídos com pool
// nil, sem precisar de banco. Os caminhos que exigem persistência estão em
// internal/app/handler/handlers_integration_test.go, contra Postgres real.
func novaAPITeste(t *testing.T) humatest.TestAPI {
	t.Helper()
	middleware.InstalarErroHuma()
	_, api := humatest.New(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := handler.NovoServicoPrecificacao(handler.Deps{Logger: logger})
	registrarOperacoes(api, svc, Repos{
		Tx:      postgres.NewTransacaoRepo(nil),
		Cotacao: postgres.NewCotacaoRepo(nil),
		Taxa:    postgres.NewTaxaBaseRepo(nil),
		Tipo:    postgres.NewTipoRecebivelRepo(nil),
	})
	return api
}

func TestSimularJSONMalformado(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Post("/api/v1/simulacoes", "Content-Type: application/json", "{invalido")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtido %d: %s", resp.Code, resp.Body.String())
	}
	var corpo dto.ErroCorpo
	if err := json.Unmarshal(resp.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("corpo não é ErroCorpo válido: %v", err)
	}
}

func TestSimularCampoObrigatorioAusente(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Post("/api/v1/simulacoes", map[string]any{})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperado 422, obtido %d: %s", resp.Code, resp.Body.String())
	}
	var corpo dto.ErroCorpo
	if err := json.Unmarshal(resp.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("corpo não é ErroCorpo válido: %v", err)
	}
	if len(corpo.Campos) == 0 {
		t.Fatal("esperada lista de campos inválidos não vazia")
	}
}

func TestSimularValorFaceDecimalInvalido(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Post("/api/v1/simulacoes", map[string]any{
		"cedente_id": 1, "tipo_recebivel": "DUPLICATA_MERCANTIL",
		"valor_face": "abc", "moeda_titulo": "BRL", "moeda_pagamento": "BRL",
		"data_vencimento": "2026-12-31",
	})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperado 422, obtido %d: %s", resp.Code, resp.Body.String())
	}
}

func TestCriarTransacaoJSONMalformado(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Post("/api/v1/transacoes", "Content-Type: application/json", "not json")
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtido %d: %s", resp.Code, resp.Body.String())
	}
}

func TestTaxaBaseVigenteSemParametroMoeda(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Get("/api/v1/taxas-base/vigente")
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperado 422, obtido %d: %s", resp.Code, resp.Body.String())
	}
}

func TestCotacaoCriarTaxaNaoPositiva(t *testing.T) {
	api := novaAPITeste(t)
	resp := api.Post("/api/v1/cotacoes", map[string]any{
		"moeda_base": "USD", "moeda_cotacao": "BRL", "taxa": "-1",
		"vigencia_inicio": "2026-01-01T00:00:00Z",
	})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperado 422, obtido %d: %s", resp.Code, resp.Body.String())
	}
}
