// Package handler implementa a camada de aplicação: orquestra o motor de
// precificação, os repositórios e o cliente de câmbio, sem conter regra de
// negócio. A regra mora no domínio; a persistência mora na infra.
package handler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/domain/money"
	"github.com/srm-asset/srm-backend/internal/domain/precificacao"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
	"github.com/srm-asset/srm-backend/internal/infra/cambio"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

type ServicoPrecificacao struct {
	motor         *precificacao.Motor
	moedas        *postgres.MoedaRepo
	tipos         *postgres.TipoRecebivelRepo
	cedentes      *postgres.CedenteRepo
	cotacoes      *postgres.CotacaoRepo
	taxas         *postgres.TaxaBaseRepo
	txs           *postgres.TransacaoRepo
	clienteCot    cambio.Cliente
	logger        *slog.Logger
}

type Deps struct {
	Motor      *precificacao.Motor
	Moedas     *postgres.MoedaRepo
	Tipos      *postgres.TipoRecebivelRepo
	Cedentes   *postgres.CedenteRepo
	Cotacoes   *postgres.CotacaoRepo
	Taxas      *postgres.TaxaBaseRepo
	Txs        *postgres.TransacaoRepo
	Cliente    cambio.Cliente
	Logger     *slog.Logger
}

func NovoServicoPrecificacao(opts Deps) *ServicoPrecificacao {
	return &ServicoPrecificacao{
		motor:      opts.Motor,
		moedas:     opts.Moedas,
		tipos:      opts.Tipos,
		cedentes:   opts.Cedentes,
		cotacoes:   opts.Cotacoes,
		taxas:      opts.Taxas,
		txs:        opts.Txs,
		clienteCot: opts.Cliente,
		logger:     opts.Logger,
	}
}

// OpcoesSimulacao agrupa os parâmetros do use case de simulação,
// atendendo ao limite S107 (≤3 params incluindo context.Context).
type OpcoesSimulacao struct {
	Req             dto.SimulacaoRequest
	Persistir       bool
}

func (s *ServicoPrecificacao) Simular(ctx context.Context, op OpcoesSimulacao) (dto.SimulacaoResponse, error) {
	entrada, err := s.prepararEntrada(ctx, op.Req)
	if err != nil {
		return dto.SimulacaoResponse{}, err
	}
	resultado, err := s.motor.Precificar(entrada)
	if err != nil {
		return dto.SimulacaoResponse{}, middleware.ClassificarErro(err)
	}
	return dto.SimulacaoResponse{
		ValorPresente:  resultado.ValorPresente8Casas.String(),
		ValorLiquido:   resultado.ValorLiquido.String(),
		Desagio:        resultado.Desagio.String(),
		MoedaTitulo:    resultado.MoedaTitulo,
		MoedaPagamento: resultado.MoedaPagamento,
		DataOperacao:   entrada.DataOperacao.Format("2006-01-02"),
		DataVencimento: entrada.DataVencimento.Format("2006-01-02"),
	}, nil
}

// OpcoesCriacao agrupa parâmetros do use case de criação de transação.
type OpcoesCriacao struct {
	Req dto.TransacaoCreateRequest
}

func (s *ServicoPrecificacao) CriarTransacao(ctx context.Context, op OpcoesCriacao) (postgres.Transacao, error) {
	entrada, err := s.prepararEntrada(ctx, dto.SimulacaoRequest(op.Req))
	if err != nil {
		return postgres.Transacao{}, err
	}
	resultado, err := s.motor.Precificar(entrada)
	if err != nil {
		return postgres.Transacao{}, middleware.ClassificarErro(err)
	}
	cedente, err := s.cedentes.PorID(ctx, op.Req.CedenteID)
	if err != nil {
		return postgres.Transacao{}, fmt.Errorf("cedente: %w", err)
	}
	tipo, err := s.tipos.PorCodigo(ctx, op.Req.TipoRecebivel)
	if err != nil {
		return postgres.Transacao{}, fmt.Errorf("tipo: %w", err)
	}
	moedaTitulo, err := s.moedas.PorCodigo(ctx, op.Req.MoedaTitulo)
	if err != nil {
		return postgres.Transacao{}, fmt.Errorf("moeda título: %w", err)
	}
	moedaPagamento, err := s.moedas.PorCodigo(ctx, op.Req.MoedaPagamento)
	if err != nil {
		return postgres.Transacao{}, fmt.Errorf("moeda pagamento: %w", err)
	}
	id, err := gerarUUID()
	if err != nil {
		return postgres.Transacao{}, err
	}
	t := postgres.Transacao{
		ID:               id,
		CedenteID:        cedente.ID,
		TipoRecebivelID:  tipo.ID,
		MoedaTituloID:    moedaTitulo.ID,
		MoedaPagamentoID: moedaPagamento.ID,
		ValorFace:        entrada.ValorFace,
		ValorPresente:    resultado.ValorPresente8Casas,
		ValorLiquido:     resultado.ValorLiquido,
		Desagio:          resultado.Desagio,
		SpreadAplicado:   resultado.SpreadAplicado,
		TaxaBaseAplicada: resultado.TaxaBaseAplicada,
		DataOperacao:     entrada.DataOperacao,
		DataVencimento:   entrada.DataVencimento,
		Status:           postgres.StatusPendente,
	}
	if resultado.CotacaoAplicada != nil {
		c := *resultado.CotacaoAplicada
		t.CotacaoAplicada = &c
	}
	if err := s.txs.Criar(ctx, t); err != nil {
		return postgres.Transacao{}, fmt.Errorf("criar transação: %w", err)
	}
	return s.txs.PorID(ctx, id)
}

func (s *ServicoPrecificacao) prepararEntrada(ctx context.Context, req dto.SimulacaoRequest) (precificacao.Entrada, error) {
	valorFace, err := money.FromString(req.ValorFace)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("entrada_invalida", "valor_face inválido", 422)
	}
	moedaTitulo, err := s.moedas.PorCodigo(ctx, req.MoedaTitulo)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("moeda_invalida", "moeda_titulo desconhecida", 422)
	}
	moedaPagamento, err := s.moedas.PorCodigo(ctx, req.MoedaPagamento)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("moeda_invalida", "moeda_pagamento desconhecida", 422)
	}
	vencimento, err := time.Parse("2006-01-02", req.DataVencimento)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("entrada_invalida", "data_vencimento inválida", 422)
	}
	dataOp := precificacao.OperacaoDataCorrente()
	if req.DataOperacao != "" {
		dataOp, err = time.Parse("2006-01-02", req.DataOperacao)
		if err != nil {
			return precificacao.Entrada{}, middleware.Novo("entrada_invalida", "data_operacao inválida", 422)
		}
	}
	_, err = s.tipos.PorCodigo(ctx, req.TipoRecebivel)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("tipo_invalido", "tipo_recebivel desconhecido", 422)
	}
	taxaBase, err := s.taxas.Vigente(ctx, moedaTitulo.Codigo, dataOp)
	if err != nil {
		return precificacao.Entrada{}, middleware.Novo("taxa_ausente", "sem taxa base vigente para a moeda na data da operação", 422)
	}
	entrada := precificacao.Entrada{
		ValorFace:      valorFace,
		DataOperacao:   dataOp,
		DataVencimento: vencimento,
		TipoRecebivel:  req.TipoRecebivel,
		TaxaBase:       taxaBase.TaxaMensal,
		MoedaTitulo:    moedaTitulo.Codigo,
		MoedaPagamento: moedaPagamento.Codigo,
	}
	if moedaTitulo.Codigo != moedaPagamento.Codigo {
		cot, err := s.cotacoes.Vigente(ctx, moedaTitulo.Codigo, moedaPagamento.Codigo, dataOp)
		if err != nil {
			return precificacao.Entrada{}, middleware.Novo("cotacao_ausente", "sem cotação vigente para o par de moedas", 422)
		}
		entrada.Cotacao = cot.Taxa
		entrada.TemCotacao = true
	}
	return entrada, nil
}

// EstrategiasDisponiveis devolve os códigos de tipo de recebível registrados
// na strategy registry. Usado pelo handler de tipos e para validação.
func EstrategiasDisponiveis(r recebivel.Registry) []string { return r.Codigos() }
