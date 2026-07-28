package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/config"
	"github.com/srm-asset/srm-backend/internal/platform/migracao"
	"github.com/srm-asset/srm-backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).
			Error("configuração inválida", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "--health" {
		checkHealth(cfg.HTTPAddr)
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := prepararBanco(ctx, cfg, logger)
	defer pool.Close()

	srv, err := server.Build(ctx, server.Deps{Cfg: cfg, Logger: logger, Pool: pool})
	if err != nil {
		logger.Error("falha ao montar servidor", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	executarComShutdownGracioso(srv, cfg, logger)
}

// prepararBanco aplica as migrations pendentes e devolve um pool pronto
// para uso. Sem isso, o schema nunca existiria e toda consulta falharia.
func prepararBanco(ctx context.Context, cfg config.Config, logger *slog.Logger) *pgxpool.Pool {
	if err := migracao.Aplicar(cfg.DatabaseURL, cfg.MigrationPath); err != nil {
		logger.Error("falha ao aplicar migrations", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	pool, err := postgres.NewPool(ctx, postgres.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    10,
		MinConns:    1,
		ConnTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Error("falha ao criar pool", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	return pool
}

// executarComShutdownGracioso inicia o servidor em background, aguarda
// SIGINT/SIGTERM e encerra dentro do prazo configurado.
func executarComShutdownGracioso(srv *http.Server, cfg config.Config, logger *slog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Info("api iniciando", slog.String("endereco", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("servidor encerrado com erro", slog.String("erro", err.Error()))
			os.Exit(1)
		}
	}()
	<-stop
	logger.Info("shutdown solicitado")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}

// checkHealth implementa o modo `--health` usado pelo healthcheck do
// Docker. A imagem é distroless (sem shell, sem curl/wget), então o
// próprio binário precisa saber se autoexaminar: faz uma requisição HTTP
// real ao endpoint /healthz do processo já em execução e sai com o
// código correspondente, sem iniciar um novo servidor.
func checkHealth(httpAddr string) {
	host, port, err := net.SplitHostPort(httpAddr)
	if err != nil || host == "" {
		host = "localhost"
	}
	url := fmt.Sprintf("http://%s/healthz", net.JoinHostPort(host, port))
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "health check falhou:", err)
		os.Exit(1)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "health check status:", resp.StatusCode)
		os.Exit(1)
	}
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
