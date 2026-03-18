package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-starter/internal/core/config"
	"go-starter/internal/core/database"
	"go-starter/internal/core/transport/http/middleware"
	"go-starter/internal/core/transport/http/server"
	auth_services "go-starter/internal/features/auth/services"
	auth_handlers "go-starter/internal/features/auth/transport/http"
	user_repository "go-starter/internal/features/user/repository"
	user_services "go-starter/internal/features/user/services"
	user_handlers "go-starter/internal/features/user/transport/http"
)

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

	// DI: Repository → Service → Handler
	userRepo := user_repository.NewUserRepository(pool)

	authHandler := auth_handlers.NewAuthHandler(auth_services.NewAuthService(userRepo, cfg), cfg)
	userHandler := user_handlers.NewUserHandler(user_services.NewUserService(userRepo))

	// Server
	srv := server.New(
		server.Config{
			Port:            cfg.AppPort,
			ShutdownTimeout: 15 * time.Second,
			DrainDelay:      5 * time.Second,
			HardStopDelay:   3 * time.Second,
		},
		logger,
		middleware.Logger(logger),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.SecurityHeaders(),
	)

	// Route groups
	authGroup := server.NewRouterGroup("/auth", middleware.RateLimiter(redisClient))
	authGroup.AddRoutes(authHandler.Routes()...)

	userGroup := server.NewRouterGroup("/user", middleware.AuthMiddleware(cfg))
	userGroup.AddRoutes(userHandler.Routes()...)

	srv.RegisterGroups(authGroup, userGroup)

	if err := srv.Run(rootCtx); err != nil {
		logger.Error("server error", slog.String("error", err.Error()))
	}
}
