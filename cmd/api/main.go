package main

import (
	"context"
	"go-starter/internal/config"
	"go-starter/internal/database"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	_shutdownPeriod      = 15 * time.Second
	_shutdownHardPeriod  = 3 * time.Second
	_readinessDrainDelay = 5 * time.Second
)

var isShuttingDown atomic.Bool

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg := config.NewConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := database.Connect(rootCtx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisClient, err := database.CreateRedisClient(cfg.RedisUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()

	handlers := SetupHandlers(cfg, pool, redisClient, logger)

	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      handlers,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		logger.Info("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	<-rootCtx.Done()
	stop()
	isShuttingDown.Store(true)
	srv.SetKeepAlivesEnabled(false)
	logger.Info("Received shutdown signal, gracefully shutting down...")

	time.Sleep(_readinessDrainDelay)
	logger.Info("Readiness check propagated, now waiting for ongoing request to finish...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownPeriod)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	stopOngoingGracefully()
	if err != nil {
		logger.Error("Failed to wait for ongoing requests to finish, waiting for forced cancellation...")
		time.Sleep(_shutdownHardPeriod)
	}
	logger.Info("Server shut down gracefully")

}
