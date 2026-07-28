package cambio

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(discard{}, nil))
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

func TestCotacaoSucesso(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"taxa":"5.4321"}`))
	}))
	defer srv.Close()

	c := NovoCliente(Opcoes{URL: srv.URL, Timeout: time.Second}, logger())
	taxa, err := c.Cotacao(context.Background(), "USD", "BRL")
	if err != nil {
		t.Fatalf("cotacao: %v", err)
	}
	if taxa != "5.4321" {
		t.Fatalf("esperado 5.4321, obtido %s", taxa)
	}
}

func TestCotacaoTaxaInvalidaRetornaErro(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"taxa":"-1"}`))
	}))
	defer srv.Close()

	c := NovoCliente(Opcoes{URL: srv.URL, Timeout: time.Second}, logger())
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
		t.Fatal("esperado erro para taxa não positiva")
	}
}

func TestCotacaoClienteDesabilitadoSemURL(t *testing.T) {
	c := NovoCliente(Opcoes{Timeout: time.Second}, logger())
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
		t.Fatal("esperado erro com URL vazia")
	}
}

// TestCircuitBreakerAbreAposFalhasConsecutivas prova as três transições de
// estado do requisito de resiliência: fechado → aberto após falhas
// consecutivas, aberto rejeita sem tentar a chamada remota, e meia-aberto
// → fechado após uma chamada de teste bem-sucedida.
func TestCircuitBreakerAbreEFecha(t *testing.T) {
	var falhar atomic.Bool
	falhar.Store(true)
	var chamadas atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas.Add(1)
		if falhar.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"taxa":"5.0"}`))
	}))
	defer srv.Close()

	c := NovoCliente(Opcoes{URL: srv.URL, Timeout: 200 * time.Millisecond, BreakerTimeout: 80 * time.Millisecond}, logger())

	// 5 falhas consecutivas abrem o circuito (ReadyToTrip: >=5).
	for i := 0; i < 5; i++ {
		if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
			t.Fatalf("chamada %d: esperada falha", i)
		}
	}
	chamadasAntesDoCircuitoAberto := chamadas.Load()

	// Circuito aberto: chamada seguinte deve falhar SEM contatar o servidor.
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
		t.Fatal("esperada falha com circuito aberto")
	}
	if chamadas.Load() != chamadasAntesDoCircuitoAberto {
		t.Fatalf("circuito aberto não deveria contatar o servidor: chamadas antes=%d depois=%d",
			chamadasAntesDoCircuitoAberto, chamadas.Load())
	}

	// Depois do BreakerTimeout o circuito vira meia-aberto e libera UMA
	// chamada de teste. O servidor volta a responder com sucesso: a
	// chamada de teste passa e o circuito fecha.
	falhar.Store(false)
	time.Sleep(120 * time.Millisecond)
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err != nil {
		t.Fatalf("chamada de teste em meia-abertura deveria suceder: %v", err)
	}

	// Circuito fechado: chamadas voltam a fluir normalmente.
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err != nil {
		t.Fatalf("circuito deveria estar fechado após sucesso em meia-abertura: %v", err)
	}
}

// TestCircuitBreakerReabreAposFalhaEmMeiaAbertura prova a quarta transição:
// se a chamada de teste em meia-abertura falha, o circuito volta a abrir em
// vez de fechar.
func TestCircuitBreakerReabreAposFalhaEmMeiaAbertura(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NovoCliente(Opcoes{URL: srv.URL, Timeout: 200 * time.Millisecond, BreakerTimeout: 80 * time.Millisecond}, logger())
	for i := 0; i < 5; i++ {
		_, _ = c.Cotacao(context.Background(), "USD", "BRL")
	}
	time.Sleep(120 * time.Millisecond)
	// Meia-aberto, mas o servidor continua falhando: a chamada de teste
	// falha e o circuito reabre.
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
		t.Fatal("esperada falha na chamada de teste em meia-abertura")
	}
	// Imediatamente aberto de novo: a chamada seguinte falha sem contatar o servidor.
	if _, err := c.Cotacao(context.Background(), "USD", "BRL"); err == nil {
		t.Fatal("esperada falha com circuito reaberto")
	}
}
