// Package server monta o servidor HTTP com Gin e injeta as dependências
// manualmente. Não há container mágico (golang-expert: DI manual é fixo).
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/srm-asset/srm-backend/internal/app/handler"
	"github.com/srm-asset/srm-backend/internal/app/middleware"
	"github.com/srm-asset/srm-backend/internal/domain/precificacao"
	"github.com/srm-asset/srm-backend/internal/domain/recebivel"
	"github.com/srm-asset/srm-backend/internal/infra/cambio"
	"github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/config"
	"github.com/srm-asset/srm-backend/internal/platform/metricas"
	"github.com/srm-asset/srm-backend/internal/report"
)

type Deps struct {
	Cfg    config.Config
	Logger *slog.Logger
	Pool   *pgxpool.Pool
}

func Build(ctx context.Context, d Deps) (*http.Server, error) {
	registry := recebivel.DefaultRegistry()
	motor := precificacao.NewMotor(registry)
	moedaRepo := postgres.NewMoedaRepo(d.Pool)
	tipoRepo := postgres.NewTipoRecebivelRepo(d.Pool)
	cedenteRepo := postgres.NewCedenteRepo(d.Pool)
	cotacaoRepo := postgres.NewCotacaoRepo(d.Pool)
	taxaRepo := postgres.NewTaxaBaseRepo(d.Pool)
	txRepo := postgres.NewTransacaoRepo(d.Pool)
	clienteCotacao := cambio.NovoCliente(cambio.Opcoes{
		URL:     d.Cfg.CambioURL,
		Timeout: time.Duration(d.Cfg.CambioTimeoutMS) * time.Millisecond,
	}, d.Logger)
	deps := handler.Deps{
		Motor: motor, Moedas: moedaRepo, Tipos: tipoRepo, Cedentes: cedenteRepo,
		Cotacoes: cotacaoRepo, Taxas: taxaRepo, Txs: txRepo,
		Cliente: clienteCotacao, Logger: d.Logger,
	}
	simulador := handler.NovoSimulador(deps)
	criador := handler.NovoCriador(deps)
	detalhe := handler.NovoDetalhe(deps)
	liquidador := handler.NovoLiquidador(deps)
	cotacaoH := handler.NovoCotacao(deps)
	taxaH := handler.NovoTaxaBase(deps)
	tiposH := handler.NovoTiposRecebivel(deps)
	extrato := report.NovoExtrato(d.Pool, d.Logger)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Correlation())
	r.Use(middleware.Recovery(d.Logger))
	r.Use(middleware.Logger(d.Logger))
	r.Use(middleware.Metricas())

	r.GET("/healthz", healthz)
	r.GET("/readyz", readyz(d.Pool))
	r.GET("/metrics", gin.WrapH(metricas.Handler()))

	v1 := r.Group(d.Cfg.APIPrefix)
	v1.POST("/simulacoes", simulador.Simular)
	v1.POST("/transacoes", criador.Criar)
	v1.GET("/transacoes", detalhe.Listar)
	v1.GET("/transacoes/:id", detalhe.Obter)
	v1.POST("/transacoes/:id/liquidacao", liquidador.Liquidar)
	v1.POST("/cotacoes", cotacaoH.Criar)
	v1.POST("/taxas-base", taxaH.Criar)
	v1.GET("/taxas-base/vigente", taxaH.Vigente)
	v1.GET("/tipos-recebivel", tiposH.Listar)
	v1.GET("/relatorios/extrato-liquidacao", extrato.Handler)

	return &http.Server{
		Addr:              d.Cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func readyz(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
