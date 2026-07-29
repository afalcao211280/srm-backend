// Package server monta o servidor HTTP com Gin e injeta as dependências
// manualmente. Não há container mágico (golang-expert: DI manual é fixo).
// As rotas de negócio são operações Huma (OpenAPI 3.1 + Swagger/docs
// automáticos) montadas sobre o mesmo *gin.Engine; healthz/readyz/metrics
// continuam gin puro, por não fazerem parte da API documentada.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
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

func Build(_ context.Context, d Deps) (*http.Server, error) {
	svc, repos := montarServico(d)
	r := montarEngine(d)

	middleware.InstalarErroHuma()
	humaConfig := huma.DefaultConfig("SRM Credit Engine API", "1.0.0")
	humaConfig.Info.Description = "Plataforma de cessão de crédito multimoedas — precificação, liquidação e extrato de recebíveis."
	api := humagin.New(r, humaConfig)

	registrarOperacoes(api, svc, repos)
	report.MontarHuma(api, d.Pool, d.Logger)

	return &http.Server{
		Addr:              d.Cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}, nil
}

func montarServico(d Deps) (*handler.ServicoPrecificacao, Repos) {
	repos := Repos{
		Tx:      postgres.NewTransacaoRepo(d.Pool),
		Cotacao: postgres.NewCotacaoRepo(d.Pool),
		Taxa:    postgres.NewTaxaBaseRepo(d.Pool),
		Tipo:    postgres.NewTipoRecebivelRepo(d.Pool),
	}
	clienteCotacao := cambio.NovoCliente(cambio.Opcoes{
		URL:     d.Cfg.CambioURL,
		Timeout: time.Duration(d.Cfg.CambioTimeoutMS) * time.Millisecond,
	}, d.Logger)
	deps := handler.Deps{
		Motor:    precificacao.NewMotor(recebivel.DefaultRegistry()),
		Moedas:   postgres.NewMoedaRepo(d.Pool),
		Tipos:    repos.Tipo,
		Cedentes: postgres.NewCedenteRepo(d.Pool),
		Cotacoes: repos.Cotacao,
		Taxas:    repos.Taxa,
		Txs:      repos.Tx,
		Cliente:  clienteCotacao,
		Logger:   d.Logger,
	}
	return handler.NovoServicoPrecificacao(deps), repos
}

func montarEngine(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Correlation())
	r.Use(middleware.Recovery(d.Logger))
	r.Use(middleware.Logger(d.Logger))
	r.Use(middleware.Metricas())

	r.GET("/healthz", healthz)
	r.GET("/readyz", readyz(d.Pool))
	r.GET("/metrics", gin.WrapH(metricas.Handler()))
	return r
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
