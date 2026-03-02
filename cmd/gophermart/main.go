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

	"github.com/bissquit/gophermart/internal/accrual"
	"github.com/bissquit/gophermart/internal/config"
	"github.com/bissquit/gophermart/internal/repository/db"
	"github.com/bissquit/gophermart/internal/server"
	"github.com/bissquit/gophermart/internal/worker"
	"github.com/bissquit/gophermart/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.GetConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := initDatabase(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	stg := db.NewDBStorage(pool, logger)
	srv := server.NewServer(cfg, stg, pool, logger)

	httpSrv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      srv.Handler(),
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	logger.Info("server starting", "addr", cfg.ServerAddr)

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddr)

	worker := worker.NewAccrualWorker(stg, accrualClient, logger)
	go worker.Start(ctx)

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}

func initDatabase(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*pgxpool.Pool, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(dbCtx, cfg.DSN)
	if err != nil {
		return nil, err
	}

	if err := migrations.InitializeDB(cfg.DSN); err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info("database connected")
	return pool, nil
}
