package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/srm-asset/srm-backend/internal/infra/postgres"
	"github.com/srm-asset/srm-backend/internal/platform/config"
	"github.com/srm-asset/srm-backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).
			Error("configuração inválida", slog.String("erro", err.Error()))
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
	defer pool.Close()

	srv, err := server.Build(ctx, server.Deps{Cfg: cfg, Logger: logger, Pool: pool})
	if err != nil {
		logger.Error("falha ao montar servidor", slog.String("erro", err.Error()))
		os.Exit(1)
	}
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
