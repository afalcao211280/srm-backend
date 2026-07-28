// Package metricas expõe os contadores e histogramas Prometheus do domínio.
// Não usa identificadores de transação ou de cedente como rótulo para manter
// a cardinalidade sob controle.
package metricas

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	Requisicoes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "srm_http_requests_total",
		Help: "Total de requisições HTTP por rota, método e status.",
	}, []string{"rota", "metodo", "status"})

	Latencia = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "srm_http_request_duration_seconds",
		Help:    "Latência por requisição HTTP.",
		Buckets: prometheus.DefBuckets,
	}, []string{"rota", "metodo", "status"})

	Precificacoes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "srm_precificacoes_total",
		Help: "Total de precificações realizadas.",
	})

	LiquidacoesSucesso = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "srm_liquidacoes_sucesso_total",
		Help: "Total de liquidações bem-sucedidas.",
	})

	LiquidacoesConflito = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "srm_liquidacoes_conflito_total",
		Help: "Total de liquidações rejeitadas por conflito de versão.",
	})

	EstadoCircuito = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "srm_circuito_cambio_estado",
		Help: "Estado do circuit breaker do cliente de câmbio (0=fechado,1=meia,2=aberto).",
	}, []string{"circuito"})
)

func init() {
	prometheus.MustRegister(Requisicoes, Latencia, Precificacoes, LiquidacoesSucesso, LiquidacoesConflito, EstadoCircuito)
}

func Handler() http.Handler { return promhttp.Handler() }
