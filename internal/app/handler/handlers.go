// Package handler concentra a camada de aplicação: o serviço de
// precificação (servico.go) e os mapeamentos DTO abaixo, reusados tanto
// pelas operações Huma (internal/server) quanto pelos testes. A ligação
// HTTP em si — bind, validação de schema, roteamento — é feita pelo Huma
// em internal/server; este pacote não depende de nenhum framework web.
package handler

import (
	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

func TransacaoParaDTO(t postgres.Transacao) dto.TransacaoResponse {
	return dto.TransacaoResponse{
		ID:             t.ID,
		Status:         string(t.Status),
		Versao:         t.Versao,
		ValorFace:      t.ValorFace.String(),
		ValorPresente:  t.ValorPresente.String(),
		ValorLiquido:   t.ValorLiquido.String(),
		Desagio:        t.Desagio.String(),
		DataOperacao:   t.DataOperacao.Format("2006-01-02"),
		DataVencimento: t.DataVencimento.Format("2006-01-02"),
		LiquidadaEm:    t.LiquidadaEm,
		CreatedAt:      t.CreatedAt,
	}
}

func CotacaoParaDTO(c postgres.Cotacao) dto.CotacaoResponse {
	return dto.CotacaoResponse{
		ID: c.ID, MoedaBase: c.MoedaBase, MoedaCotacao: c.MoedaCotacao,
		Taxa: c.Taxa.String(), VigenciaInicio: c.VigenciaInicio, VigenciaFim: c.VigenciaFim,
	}
}

func TaxaBaseParaDTO(t postgres.TaxaBase) dto.TaxaBaseResponse {
	return dto.TaxaBaseResponse{
		ID: t.ID, Moeda: t.Moeda, TaxaMensal: t.TaxaMensal.String(),
		VigenciaInicio: t.VigenciaInicio, VigenciaFim: t.VigenciaFim,
	}
}
