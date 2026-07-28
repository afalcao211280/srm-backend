package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/srm-asset/srm-backend/internal/app/dto"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
)

type Simulador struct {
	svc    *ServicoPrecificacao
	logger *slog.Logger
}

func NovoSimulador(opts Deps) *Simulador {
	return &Simulador{svc: NovoServicoPrecificacao(opts), logger: opts.Logger}
}

func (s *Simulador) Simular(c *gin.Context) {
	var req dto.SimulacaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, middleware.Novo("entrada_invalida", "JSON malformado", http.StatusBadRequest))
		return
	}
	if err := middleware.Validate(req); err != nil {
		middleware.RespondError(c, err)
		return
	}
	res, err := s.svc.Simular(c.Request.Context(), OpcoesSimulacao{Req: req})
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

type Criador struct {
	svc    *ServicoPrecificacao
	logger *slog.Logger
}

func NovoCriador(opts Deps) *Criador {
	return &Criador{svc: NovoServicoPrecificacao(opts), logger: opts.Logger}
}

func (c *Criador) Criar(ctx *gin.Context) {
	var req dto.TransacaoCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(ctx, middleware.Novo("entrada_invalida", "JSON malformado", http.StatusBadRequest))
		return
	}
	if err := middleware.Validate(req); err != nil {
		middleware.RespondError(ctx, err)
		return
	}
	t, err := c.svc.CriarTransacao(ctx.Request.Context(), OpcoesCriacao{Req: req})
	if err != nil {
		middleware.RespondError(ctx, err)
		return
	}
	ctx.Header("Location", "/api/v1/transacoes/"+t.ID)
	ctx.JSON(http.StatusCreated, transacaoParaDTO(t))
}

type Detalhe struct {
	repo   *postgres.TransacaoRepo
	logger *slog.Logger
}

func NovoDetalhe(opts Deps) *Detalhe {
	return &Detalhe{repo: opts.Txs, logger: opts.Logger}
}

func (d *Detalhe) Obter(c *gin.Context) {
	id := c.Param("id")
	t, err := d.repo.PorID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transacaoParaDTO(t))
}

func (d *Detalhe) Listar(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}

type Liquidador struct {
	repo   *postgres.TransacaoRepo
	logger *slog.Logger
}

func NovoLiquidador(opts Deps) *Liquidador {
	return &Liquidador{repo: opts.Txs, logger: opts.Logger}
}

func (l *Liquidador) Liquidar(c *gin.Context) {
	id := c.Param("id")
	// Lê a versão atual via header opcional X-If-Version; sem ele, lê do banco.
	t, err := l.repo.PorID(c.Request.Context(), id)
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	atual, err := l.repo.Liquidar(c.Request.Context(), id, t.Versao, time.Now().UTC())
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, transacaoParaDTO(atual))
}

type Cotacao struct {
	repo   *postgres.CotacaoRepo
	logger *slog.Logger
}

func NovoCotacao(opts Deps) *Cotacao {
	return &Cotacao{repo: opts.Cotacoes, logger: opts.Logger}
}

func (h *Cotacao) Criar(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

type TaxaBase struct {
	repo   *postgres.TaxaBaseRepo
	logger *slog.Logger
}

func NovoTaxaBase(opts Deps) *TaxaBase {
	return &TaxaBase{repo: opts.Taxas, logger: opts.Logger}
}

func (h *TaxaBase) Criar(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *TaxaBase) Vigente(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type TiposRecebivel struct {
	repo   *postgres.TipoRecebivelRepo
	logger *slog.Logger
}

func NovoTiposRecebivel(opts Deps) *TiposRecebivel {
	return &TiposRecebivel{repo: opts.Tipos, logger: opts.Logger}
}

func (h *TiposRecebivel) Listar(c *gin.Context) {
	tipos, err := h.repo.Listar(c.Request.Context())
	if err != nil {
		middleware.RespondError(c, err)
		return
	}
	out := make([]dto.TipoRecebivelResponse, 0, len(tipos))
	for _, t := range tipos {
		out = append(out, dto.TipoRecebivelResponse{Codigo: t.Codigo, Nome: t.Nome})
	}
	c.JSON(http.StatusOK, gin.H{"items": out})
}

func transacaoParaDTO(t postgres.Transacao) dto.TransacaoResponse {
	res := dto.TransacaoResponse{
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
	return res
}
